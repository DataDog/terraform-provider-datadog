#!/usr/bin/env python3
"""
Reads the test2json stream gotestsum writes via --jsonfile and emits a check-run
output ({"title": ..., "summary": ...}) describing the acceptance-test results.

The PR integration workflow surfaces this on the pull request, which the
workflow_run-triggered job cannot do by posting job status alone.

Usage:
    python3 scripts/summarize_testacc.py --jsonfile /tmp/testacc.json \
        [--selected-regex "^TestA$|^TestB$"] [--skipped-regex "^TestB$"] \
        [--details-url URL]

A test the -run regex selected but that never appears in the stream was either
held back by the -skip list or reported nothing at all; --skipped-regex is what
tells those apart. A test that failed and then passed was rerun by gotestsum
--rerun-fails; it is reported as flaky but does not fail the check, which mirrors
gotestsum's own exit code.
"""

import argparse
import json
import os
import sys
from typing import Dict, List, Optional

_TERMINAL_ACTIONS = ("pass", "fail", "skip")


def _selected_names(regex: Optional[str]) -> List[str]:
    """Recover test names from the anchored, pipe-joined regex we generate.

    select_pr_tests.py only ever emits `^Name$` alternatives, so this is exact
    rather than a general regex parse. Anything unexpected is ignored.
    """
    if not regex:
        return []
    names = []
    for part in regex.split("|"):
        part = part.strip()
        if part.startswith("^") and part.endswith("$"):
            names.append(part[1:-1])
    return names


def parse_results(stream: str) -> Dict[str, List[str]]:
    """Map top-level test name -> terminal actions, in the order they occurred.

    Subtests (names containing "/") are folded into their parent, which already
    reports the aggregate result.
    """
    results: Dict[str, List[str]] = {}
    for line in stream.splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        name = event.get("Test")
        action = event.get("Action")
        if not name or "/" in name or action not in _TERMINAL_ACTIONS:
            continue
        results.setdefault(name, []).append(action)
    return results


def summarize(
    stream: str,
    *,
    selected_regex: Optional[str] = None,
    skipped_regex: Optional[str] = None,
    details_url: Optional[str] = None,
) -> Dict[str, str]:
    results = parse_results(stream)
    selected = _selected_names(selected_regex)
    skip_listed = set(_selected_names(skipped_regex))

    failed, flaky, passed, skipped = [], [], [], []
    for name, actions in sorted(results.items()):
        final = actions[-1]
        if final == "fail":
            failed.append(name)
        elif final == "skip":
            skipped.append(name)
        elif "fail" in actions:
            flaky.append(name)
        else:
            passed.append(name)

    # A selected test with no result was either held back by the flaky-test skip
    # list or never reported one, which is what a compile error or an aborted run
    # looks like. Those are different problems, so don't collapse them.
    ran = set(results)
    filtered = sorted(n for n in selected if n not in ran and n in skip_listed)
    missing = sorted(n for n in selected if n not in ran and n not in skip_listed)

    counts = []
    if failed:
        counts.append(f"{len(failed)} failed")
    if missing:
        counts.append(f"{len(missing)} without a result")
    if flaky:
        counts.append(f"{len(flaky)} flaky")
    if passed:
        counts.append(f"{len(passed)} passed")
    if skipped:
        counts.append(f"{len(skipped)} skipped")
    if filtered:
        counts.append(f"{len(filtered)} not run")
    title = ", ".join(counts) if counts else "No tests ran"

    header = []
    if selected:
        header.append(f"{len(selected)} test(s) selected by the changed files in this PR.")
        header.append("")
    header.append("| Test | Result |")
    header.append("| --- | --- |")

    rows = []
    for name in failed:
        rows.append(f"| `{name}` | :x: failed |")
    for name in missing:
        rows.append(f"| `{name}` | :grey_question: no result reported |")
    for name in flaky:
        rows.append(f"| `{name}` | :warning: flaky — failed, passed on re-run |")
    for name in filtered:
        rows.append(f"| `{name}` | :fast_forward: not run — in `flaky_tests.yaml` |")
    for name in skipped:
        rows.append(f"| `{name}` | :heavy_minus_sign: skipped |")
    for name in passed:
        rows.append(f"| `{name}` | :white_check_mark: passed |")

    footer = []
    if missing:
        footer.append("")
        footer.append(
            "A test with no result was selected to run but reported neither pass "
            "nor fail, which usually means the package failed to build or the run "
            "was cut short."
        )
    if flaky:
        footer.append("")
        footer.append(
            "Flaky tests do not fail this check: `gotestsum --rerun-fails` reruns "
            "them and exits on the final result."
        )
    if details_url:
        footer.append("")
        footer.append(f"[Full logs]({details_url})")

    # The check-run output.summary field is capped at 65535 characters. Exceeding
    # it makes the PATCH 422 and leaves the check stuck in progress, which blocks
    # merges once it is required, so drop rows rather than the whole update. Rows
    # are already ordered worst-first, so truncation loses the least useful ones.
    rows = _fit(header, rows, footer)

    return {"title": title, "summary": "\n".join(header + rows + footer)}


_MAX_SUMMARY_CHARS = 65535


def _fit(header: List[str], rows: List[str], footer: List[str]) -> List[str]:
    """Trim rows until header + rows + notice + footer fits the check-run limit."""

    def size(body: List[str]) -> int:
        return len("\n".join(header + body + footer))

    if size(rows) <= _MAX_SUMMARY_CHARS:
        return rows

    kept = list(rows)
    while kept:
        notice = ["", f"_Showing {len(kept)} of {len(rows)} tests; see the full logs._"]
        if size(kept + notice) <= _MAX_SUMMARY_CHARS:
            return kept + notice
        # Drop proportionally at first so this converges quickly on large suites.
        kept = kept[: -max(1, len(kept) // 10)]
    return ["", f"_All {len(rows)} test results were too large to show; see the full logs._"]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--jsonfile", required=True)
    parser.add_argument("--selected-regex", default=None)
    parser.add_argument("--skipped-regex", default=None)
    parser.add_argument("--details-url", default=None)
    args = parser.parse_args()

    if not os.path.isfile(args.jsonfile):
        print(
            json.dumps(
                {
                    "title": "No test results",
                    "summary": "The acceptance-test run produced no results file.",
                }
            )
        )
        return

    with open(args.jsonfile, errors="replace") as fh:
        stream = fh.read()

    json.dump(
        summarize(
            stream,
            selected_regex=args.selected_regex,
            skipped_regex=args.skipped_regex,
            details_url=args.details_url,
        ),
        sys.stdout,
    )
    print()


if __name__ == "__main__":
    main()
