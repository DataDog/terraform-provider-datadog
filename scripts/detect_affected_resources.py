#!/usr/bin/env python3
"""
Detects which resources and data sources an issue or pull request is about and
prints the label changes to make, one per line as "add\t<label>\t<description>"
or "remove\t<label>".

Usage:
    R=repos/DataDog/terraform-provider-datadog
    gh api $R/issues/123 > item.json
    gh api --paginate $R/issues/123/events \
        --jq '.[] | select(.event == "labeled" or .event == "unlabeled")
              | {event, actor: .actor.login, label: .label.name} | tojson' > events.jsonl
    python3 scripts/detect_affected_resources.py --item item.json --events events.jsonl

    # Pull requests also need their changed files and how many there are
    gh api --paginate $R/pulls/123/files --jq '.[].filename' > files.txt
    python3 scripts/detect_affected_resources.py --item item.json --events events.jsonl \
        --files files.txt --changed-files "$(gh api $R/pulls/123 --jq .changed_files)"

Labels are only removed when github-actions[bot] added them and they no longer
apply. Labels people add are never removed, and labels people remove are never
added back.

Signals:
    - Issues: the title, and the "What resources or data sources are affected?"
      section of the issue form. Issues without that section fall back to the
      body, ignoring fenced code blocks (pasted configs mention many resources).
    - Pull requests: the title ("[datadog_xxx] ...") and the changed files.

Only names documented under docs/resources and docs/data-sources are labeled.
Nothing is labeled when the item looks provider-wide: more than MAX_UNRELATED
names that no single CODEOWNERS team owns, or more than MAX_LABELS names.

This script only reads text and file names. It is run by a pull_request_target
workflow, so it must never execute anything taken from the pull request.

Exit behavior:
    - Empty output means no changes
    - Reasons for skipping are written to stderr
    - Changes nothing if the PR's file list is incomplete (the API lists at
      most 3,000 files)
    - Exits non-zero, changing nothing, if the docs can't be found
"""

import argparse
import hashlib
import json
import os
import re
import sys
from typing import Dict, Iterable, List, Optional, Set, Tuple

RESOURCE = "resource"
DATASOURCE = "datasource"
_LABEL_PREFIX = {RESOURCE: "resource/", DATASOURCE: "datasource/"}
_DOCS_DIR = {RESOURCE: "resources", DATASOURCE: "data-sources"}

# GitHub rejects longer label names.
MAX_LABEL_LEN = 50
# Longer labels end in "-" plus this many hex digits of a hash of the full
# label. Names never contain "-", so these can't clash with unshortened labels.
_HASH_LEN = 6
# Above this many names, only label when one team owns all of them.
MAX_UNRELATED = 3
# Never add more labels than this, even when one team owns everything.
MAX_LABELS = 10
# Only scan the start of very long bodies.
MAX_TEXT_LEN = 64 * 1024

# Who adds labels when the workflow uses GITHUB_TOKEN. Only labels this
# account added are ever removed.
BOT_LOGIN = "github-actions[bot]"

# Owners that appear on almost every path and say nothing about relatedness.
_DEFAULT_OWNERS = {"@datadog/api-reliability", "@datadog/documentation"}

_AFFECTED_HEADING_RE = re.compile(
    r"^###\s+What resources or data sources are affected\?\s*$", re.MULTILINE
)
_NEXT_HEADING_RE = re.compile(r"^###\s", re.MULTILINE)
_CODE_BLOCK_RE = re.compile(r"(`{3,}|~{3,}).*?(?:\1|\Z)", re.DOTALL)

# Wording that says the item is about the provider as a whole. Names given as
# examples next to it ("All resources, e.g. datadog_monitor") don't count.
# Explicit markers count everywhere, including tagged titles.
_PROVIDER_MARKER_RE = re.compile(
    r"\bprovider[ -](?:wide|level)\b"
    r"|\bthe\s+provider\s+itself\b"
    r"|\bterraform(?:'s)?\s+core\b"
    r"|^\s*(?:\[[^\]]*\]\s*)*\[provider\]",
    re.IGNORECASE,
)
_PROVIDER_WIDE_RE = re.compile(
    # Plural after "all", singular after "every"/"any", so phrases like
    # "all resource metric types" don't count. "Multiple"/"several"/"many"
    # usually count instances ("multiple resources with for_each"), and
    # several distinct names are handled by is_clear instead.
    r"\ball\s+(?:of\s+the\s+)?(?:datadog\s+)?(?:resources|data[ _-]?sources)\b"
    r"|\b(?:every|any)\s+(?:datadog\s+)?(?:resource|data[ _-]?source)\b"
    r"|" + _PROVIDER_MARKER_RE.pattern,
    re.IGNORECASE,
)
# Negation just before provider-wide wording: "not provider-wide",
# "doesn't affect all resources". At most one word in between, so "not just
# datadog_x but all resources" still counts.
_NEGATION_BEFORE_RE = re.compile(r"(?:\b(?:not|no|never|cannot)|n['’]t)\s+(?:[\w'’-]+\s+)?$", re.IGNORECASE)
# A title like "[datadog_monitor] ..." or "[ACTP][datadog_x] ..." already
# names its scope, so "all"/"every resource" later in it ("every resource is
# recreated") is about instances. Only _PROVIDER_MARKER_RE applies to it.
_TITLE_TAG_RE = re.compile(r"^\s*(?:\[[^\]]*\]\s*)*?\[(?:data\.)?datadog_[a-z0-9_]+\]")

# Lowercase only, so env vars such as DATADOG_API_KEY are not read as names.
# A preceding dot is allowed for module addresses such as
# "module.x.datadog_monitor.y" and "module.x.data.datadog_monitor.y", but not
# right after "module." (module names like "module.datadog_alerts").
_TEXT_NAME_RE = re.compile(
    r"(?<![A-Za-z0-9_])(?<!module\.)(data\.|data_source_|resource_)?(datadog_[a-z0-9_]+)(?![A-Za-z0-9])"
)
_HCL_DATA_RE = re.compile(r'\bdata\s+"$')
# Kind words right before a name, possibly both: "data source datadog_x",
# "resource and data source datadog_x". The word next to the name must be
# singular for resources: "Resources:" headings are handled by
# _KIND_HEADING_RE, and "Affected resources: datadog_x" says nothing about the
# kind.
_BEFORE_KINDS_RE = re.compile(
    r"\b((?:(?:data[ _-]?sources?|resources?)\s*(?:and|or|&|/|,)\s*)*(?:data[ _-]?sources?|resource))\W{0,3}$",
    re.IGNORECASE,
)
# Kind words right after a name, possibly both: "datadog_x data source",
# "datadog_x (data source and resource)".
_KIND_WORD_RE = re.compile(r"data[ _-]?sources?|resources?", re.IGNORECASE)
_AFTER_KINDS_RE = re.compile(
    r"^\W{0,3}((?:data[ _-]?sources?|resources?)(?:\s*(?:and|or|&|/|,)\s*(?:data[ _-]?sources?|resources?))*)\b",
    re.IGNORECASE,
)
# What may separate names that share the words around them.
_NAME_JOIN_RE = re.compile(r"[\s`'\"*,]*(?:(?:and|or|&|/)[\s`'\"*,]*)?", re.IGNORECASE)
# A line that is only a kind heading, like "Data sources:", "## Resources",
# "**Resources**" or "- Data source: datadog_x, datadog_y". Group 2 is the rest
# of the line.
_KIND_HEADING_RE = re.compile(
    r"^\s*(?:#{1,6}\s+|[-*>]\s*)?[*_]*(data[ _-]?sources?|resources?)[*_]*\s*(?::[*_]*\s*(.*))?$", re.IGNORECASE
)
_MARKDOWN_HEADING_RE = re.compile(r"^\s*#{1,6}\s")

_SOURCE_FILE_RE = re.compile(
    r"^datadog/(?:fwprovider/|tests/)?(resource|data_source)_(?:datadog_)?(\w+?)(_test)?\.go$"
)
_DOCS_FILE_RE = re.compile(r"^(?:docs|templates)/(resources|data-sources)/(\w+)\.md(?:\.tmpl)?$")
_EXAMPLES_FILE_RE = re.compile(r"^examples/(resources|data-sources)/datadog_(\w+)/")
_NEW_DOCS_FILE_RE = re.compile(r"^docs/(resources|data-sources)/(\w+)\.md$")
_TRUSTED_AUTHORS = {"OWNER", "MEMBER", "COLLABORATOR"}
_FILE_KIND = {
    "resource": RESOURCE,
    "data_source": DATASOURCE,
    "resources": RESOURCE,
    "data-sources": DATASOURCE,
}
# Directories that belong to specific resources but don't carry their names.
_DIR_NAMES = {
    "datadog/dashboardmapping/": [(RESOURCE, "datadog_dashboard_v2"), (RESOURCE, "datadog_powerpack_v2")],
    "datadog/fwprovider/observability_pipeline/": [(RESOURCE, "datadog_observability_pipeline")],
}

Known = Dict[str, Set[str]]
Name = Tuple[str, str]  # (kind, "datadog_xxx")


def load_known_names(repo_root: str) -> Known:
    known: Known = {}
    for kind, subdir in _DOCS_DIR.items():
        try:
            files = os.listdir(os.path.join(repo_root, "docs", subdir))
        except FileNotFoundError:
            files = []
        known[kind] = {"datadog_" + f[: -len(".md")] for f in files if f.endswith(".md")}
    return known


def _pick_kind(name: str, preferred: Optional[str], known: Known) -> Optional[str]:
    if not any(name in names for names in known.values()):
        return None
    if preferred:
        return preferred
    for kind in (RESOURCE, DATASOURCE):
        if name in known[kind]:
            return kind
    return None


def names_from_text(text: str, known: Known) -> Dict[Name, bool]:
    """Return the names found, each mapped to whether the text stated its kind.

    In order, the kind is stated by:
      - the name's own prefix: "data.datadog_x", "resource_datadog_x";
      - the words right before it: 'data "datadog_x"', "data source datadog_x",
        "resource datadog_x";
      - the words right after it, possibly both kinds: "datadog_x data
        source", "datadog_x (data source and resource)";
      - the latest heading, such as "Data sources:" or "Resources".
    Names joined only by "and", "or", commas or slashes share the words around
    them, so "datadog_a and datadog_b data sources" are both data sources.
    """
    found: Dict[Name, bool] = {}
    heading: Optional[str] = None

    def add(name: str, preferred: Optional[str]) -> None:
        kind = _pick_kind(name, preferred, known)
        if kind:
            found[(kind, name)] = found.get((kind, name), False) or kind == preferred

    for line in text[:MAX_TEXT_LEN].splitlines():
        if _MARKDOWN_HEADING_RE.match(line):
            heading = None
        h = _KIND_HEADING_RE.match(line)
        if h:
            heading = DATASOURCE if h.group(1).lower().startswith("data") else RESOURCE
            line = h.group(2) or ""

        matches = list(_TEXT_NAME_RE.finditer(line))
        groups: List[List["re.Match[str]"]] = []
        for i, m in enumerate(matches):
            if i and _NAME_JOIN_RE.fullmatch(line[matches[i - 1].end() : m.start()]):
                groups[-1].append(m)
            else:
                groups.append([m])

        for group in groups:
            start = group[0].start()
            before = _BEFORE_KINDS_RE.search(line, max(0, start - 40), start)
            after = _AFTER_KINDS_RE.match(line[group[-1].end() : group[-1].end() + 40])
            if _HCL_DATA_RE.search(line, max(0, start - 20), start):
                stated = [DATASOURCE]
            elif before:
                stated = sorted({_kind_of_word(w) for w in _KIND_WORD_RE.findall(before.group(1))})
            elif after:
                stated = sorted({_kind_of_word(w) for w in _KIND_WORD_RE.findall(after.group(1))})
            else:
                stated = [heading] if heading else [None]
            for m in group:
                prefix, name = m.group(1), m.group(2).rstrip("_")
                if prefix in ("data.", "data_source_"):
                    add(name, DATASOURCE)
                elif prefix == "resource_":
                    add(name, RESOURCE)
                else:
                    for preferred in stated:
                        add(name, preferred)
    return found


def _kind_of_word(word: str) -> str:
    return DATASOURCE if word.lower().startswith("data") else RESOURCE


def _prefer_explicit(*sources: Dict[Name, bool]) -> Set[Name]:
    """Merge names, dropping guessed kinds for names whose kind was stated.

    E.g. "[datadog_monitor] ..." guesses a resource, but "data.datadog_monitor"
    elsewhere states a data source, so only the data source is kept.
    """
    merged: Dict[Name, bool] = {}
    for source in sources:
        for n, explicit in source.items():
            merged[n] = merged.get(n, False) or explicit
    stated = {name for (_, name), explicit in merged.items() if explicit}
    return {(kind, name) for (kind, name), explicit in merged.items() if explicit or name not in stated}


def _longest_prefix(suffix: str, candidates: Set[str]) -> Optional[str]:
    """Map e.g. "dashboard_heatmap_test" to "datadog_dashboard" if a candidate."""
    parts = suffix.split("_")
    for i in range(len(parts), 0, -1):
        name = "datadog_" + "_".join(parts[:i])
        if name in candidates:
            return name
    return None


def names_from_files(files: Iterable[str], known: Known) -> Set[Name]:
    files = [f.strip() for f in files]
    found: Set[Name] = set()
    # Docs, examples and non-test source files are named exactly after their
    # resource. Names they introduce that aren't known stay unlabeled, and
    # also stop test files of that new resource from falling back to a
    # shorter, unrelated name (..._monitor_new_feature_test.go is not
    # datadog_monitor), whether or not the docs exist yet.
    unknown: Known = {kind: set() for kind in known}
    tests: List[Tuple[str, str]] = []
    for f in files:
        for prefix, names in _DIR_NAMES.items():
            if f.startswith(prefix):
                found.update(n for n in names if n[1] in known[n[0]])
        m = _SOURCE_FILE_RE.match(f)
        if m and m.group(3):
            tests.append((_FILE_KIND[m.group(1)], m.group(2)))
            continue
        m = m or _DOCS_FILE_RE.match(f) or _EXAMPLES_FILE_RE.match(f)
        if not m:
            continue
        # resource_datadog_synthetics_test_.go ends in "_" so Go doesn't
        # treat it as a test file.
        kind, name = _FILE_KIND[m.group(1)], "datadog_" + m.group(2).rstrip("_")
        if name in known[kind]:
            found.add((kind, name))
        else:
            unknown[kind].add(name)

    # Test files may add a suffix to the resource name, e.g.
    # resource_datadog_dashboard_heatmap_test.go.
    for kind, suffix in tests:
        name = _longest_prefix(suffix, known[kind] | unknown[kind])
        if name in known[kind]:
            found.add((kind, name))
    return found


def _is_provider_wide(text: str, pattern: "re.Pattern[str]" = _PROVIDER_WIDE_RE) -> bool:
    return any(
        not _NEGATION_BEFORE_RE.search(text, max(0, m.start() - 40), m.start())
        for m in pattern.finditer(text)
    )


def affected_section(body: str) -> Optional[str]:
    """Return the issue form's "affected resources" answer, if present."""
    m = _AFFECTED_HEADING_RE.search(body)
    if not m:
        return None
    rest = body[m.end() :]
    for nxt in _NEXT_HEADING_RE.finditer(rest):
        # "### Data sources" inside the answer is part of it, not the next field.
        line = rest[nxt.start() :].split("\n", 1)[0]
        if not _KIND_HEADING_RE.match(line):
            return rest[: nxt.start()]
    return rest


def _codeowners_rules(text: str) -> List[Tuple["re.Pattern[str]", Set[str]]]:
    rules = []
    for line in text.splitlines():
        line = line.split("#", 1)[0].strip()
        if not line:
            continue
        pattern, *owners = line.split()
        anchored = "/" in pattern.rstrip("/")
        pattern = pattern.lstrip("/")
        if pattern.endswith("/"):
            pattern += "**"
        regex = ""
        i = 0
        while i < len(pattern):
            if pattern.startswith("**/", i):
                regex, i = regex + "(?:.*/)?", i + 3
            elif pattern.startswith("**", i):
                regex, i = regex + ".*", i + 2
            elif pattern[i] == "*":
                regex, i = regex + "[^/]*", i + 1
            elif pattern[i] == "?":
                regex, i = regex + "[^/]", i + 1
            else:
                regex, i = regex + re.escape(pattern[i]), i + 1
        if not anchored:
            regex = "(?:.*/)?" + regex
        rules.append((re.compile(regex + "(?:/.*)?$"), {o.lower() for o in owners}))
    return rules


def load_owners(repo_root: str):
    """Return a function mapping a name to the teams that own it in CODEOWNERS."""
    try:
        with open(os.path.join(repo_root, ".github", "CODEOWNERS")) as fh:
            rules = _codeowners_rules(fh.read())
    except OSError:
        rules = []

    def owners_of(name: str) -> Set[str]:
        short = name[len("datadog_") :]
        probes = [
            f"datadog/fwprovider/resource_{name}.go",
            f"datadog/resource_{name}.go",
            f"docs/resources/{short}.md",
            f"examples/resources/{name}/resource.tf",
        ]
        owners: Set[str] = set()
        for path in probes:
            # Like GitHub, the last matching rule wins for each path.
            for regex, rule_owners in reversed(rules):
                if regex.match(path):
                    owners |= rule_owners
                    break
        return owners - _DEFAULT_OWNERS

    return owners_of


def is_clear(names: Set[Name], owners_of) -> bool:
    bases = {name for _, name in names}
    if len(bases) <= MAX_UNRELATED:
        return True
    if len(names) > MAX_LABELS:
        return False
    common: Optional[Set[str]] = None
    for name in bases:
        owners = owners_of(name)
        common = owners if common is None else common & owners
    return bool(common)


def label_for(kind: str, name: str) -> str:
    label = _LABEL_PREFIX[kind] + name
    if len(label) <= MAX_LABEL_LEN:
        return label
    # Plain truncation would merge names sharing their first 50 characters,
    # e.g. ..._ticket_creation_rule and ..._ticket_creation_rules_order.
    digest = hashlib.sha256(label.encode()).hexdigest()[:_HASH_LEN]
    return f"{label[: MAX_LABEL_LEN - _HASH_LEN - 1]}-{digest}"


def detect_labels(item: dict, files: List[str], repo_root: str) -> List[Tuple[str, str]]:
    """Return sorted (label, description) pairs to add to the issue or PR."""
    known = load_known_names(repo_root)
    title = item.get("title") or ""
    body = item.get("body") or ""

    if "pull_request" in item and item.get("author_association") in _TRUSTED_AUTHORS:
        # A PR adding a new resource documents it in the same PR. Only trusted
        # authors may introduce names, so fork PRs can't create arbitrary labels.
        for f in files:
            m = _NEW_DOCS_FILE_RE.match(f.strip())
            if m:
                known[_FILE_KIND[m.group(1)]].add("datadog_" + m.group(2))

    title_pattern = _PROVIDER_MARKER_RE if _TITLE_TAG_RE.match(title) else _PROVIDER_WIDE_RE
    scope = [(title, title_pattern)]
    from_title = names_from_text(title, known)
    if "pull_request" in item:
        # File paths always state the kind.
        from_files = dict.fromkeys(names_from_files(files, known), True)
        names = _prefer_explicit(from_title, from_files)
    else:
        section = affected_section(body)
        if section is None:
            section = _CODE_BLOCK_RE.sub("", body)
        # Pasted configs can't make an issue provider-wide.
        scope.append((_CODE_BLOCK_RE.sub("", section), _PROVIDER_WIDE_RE))
        names = _prefer_explicit(from_title, names_from_text(section, known))

    if any(_is_provider_wide(text[:MAX_TEXT_LEN], pattern) for text, pattern in scope):
        print("Provider-wide, leaving unlabeled", file=sys.stderr)
        return []

    if not names:
        print("No known resource or data source found", file=sys.stderr)
        return []
    if not is_clear(names, load_owners(repo_root)):
        print(
            "Too many unrelated names, leaving unlabeled: "
            + ", ".join(sorted(f"{k}/{n}" for k, n in names)),
            file=sys.stderr,
        )
        return []

    labels = {
        label_for(kind, name): f"{'Data source' if kind == DATASOURCE else 'Resource'} {name}"
        for kind, name in names
    }
    return sorted(labels.items())


def _is_ours(label: str) -> bool:
    return label.startswith(tuple(_LABEL_PREFIX.values()))


def plan_label_changes(
    item: dict,
    files: List[str],
    events: List[dict],
    repo_root: str,
    changed_files: Optional[int] = None,
) -> Tuple[List[Tuple[str, str]], List[str]]:
    """Return (labels to add as (label, description), labels to remove).

    events are the item's "labeled"/"unlabeled" events, oldest first, as
    {"event", "actor", "label"}. The latest event for a label says who owns it:
    only labels BOT_LOGIN added are ever removed, and labels a person removed
    are never added back.

    changed_files is the PR's changed_files count. When files has a different
    number of entries, nothing changes.
    """
    if not any(load_known_names(repo_root).values()):
        # Without the docs every label would look stale; never act on that.
        raise SystemExit("No docs/resources or docs/data-sources found")
    if changed_files is not None and len(files) != changed_files:
        # GET /pulls/{n}/files returns at most 3,000 files. Labels for the
        # missing files would look stale, so don't act on a partial list.
        # A longer list also indicates inconsistent reads, e.g. a push
        # between fetching PR metadata and fetching its files.
        print(f"Listed {len(files)} changed files, expected {changed_files}; changing nothing", file=sys.stderr)
        return [], []
    if changed_files is not None and len(set(files)) != len(files):
        # Overlapping pages must not count the same path twice and hide a
        # missing file, even when the total count happens to match.
        print("Duplicate changed file paths listed; changing nothing", file=sys.stderr)
        return [], []

    last: Dict[str, dict] = {}
    for event in events:
        if event.get("event") in ("labeled", "unlabeled"):
            last[event.get("label") or ""] = event
    added_by_bot = {lbl for lbl, e in last.items() if e["event"] == "labeled" and e.get("actor") == BOT_LOGIN}
    removed_by_person = {lbl for lbl, e in last.items() if e["event"] == "unlabeled" and e.get("actor") != BOT_LOGIN}

    wanted = dict(detect_labels(item, files, repo_root))
    current = {lbl.get("name") for lbl in item.get("labels") or []}
    to_add = sorted((lbl, desc) for lbl, desc in wanted.items() if lbl not in current | removed_by_person)
    to_remove = sorted(lbl for lbl in current if _is_ours(lbl) and lbl in added_by_bot and lbl not in wanted)
    for lbl in sorted(set(wanted) & removed_by_person - current):
        print(f"Not re-adding {lbl}: removed by a person", file=sys.stderr)
    if not to_add and not to_remove:
        print("Labels already up to date", file=sys.stderr)
    return to_add, to_remove


def _read_events(path: str) -> List[dict]:
    """Read events as JSON lines, as gh api --paginate --jq '.[]' prints them."""
    with open(path) as fh:
        return [json.loads(line) for line in fh if line.strip()]


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[1])
    parser.add_argument("--item", required=True, help="issue or PR JSON from GET /repos/{repo}/issues/{n}")
    parser.add_argument("--files", help="changed file paths, one per line (PRs only)")
    parser.add_argument("--events", help="label events as JSON lines of {event, actor, label}, oldest first")
    parser.add_argument(
        "--changed-files", type=int, help="changed_files from GET /repos/{repo}/pulls/{n} (required for PRs)"
    )
    args = parser.parse_args()

    with open(args.item) as fh:
        item = json.load(fh)
    if item.get("pull_request") and args.changed_files is None:
        parser.error("--changed-files is required for pull requests")
    files: List[str] = []
    if args.files:
        with open(args.files) as fh:
            files = [line for line in fh.read().splitlines() if line.strip()]
    events = _read_events(args.events) if args.events else []

    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    to_add, to_remove = plan_label_changes(item, files, events, repo_root, args.changed_files)
    for label, description in to_add:
        print(f"add\t{label}\t{description}")
    for label in to_remove:
        print(f"remove\t{label}")


if __name__ == "__main__":
    main()
