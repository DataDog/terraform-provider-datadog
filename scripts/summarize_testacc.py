#!/usr/bin/env python3
"""
Reads the test2json stream gotestsum writes via --jsonfile and emits a check-run
output ({"title": ..., "summary": ...}) describing the acceptance-test results.

The PR integration workflow surfaces this on the pull request, which the
workflow_run-triggered job cannot do by posting job status alone.

Usage:
    python3 scripts/summarize_testacc.py --jsonfile /tmp/testacc.json \
        [--selected-regex "^TestA$|^TestB$"] [--details-url URL]

A test the -run regex selected but that never appears in the stream was filtered
out by the flaky-test -skip list. A test that failed and then passed was rerun by
gotestsum --rerun-fails; it is reported as flaky but does not fail the check,
which mirrors gotestsum's own exit code.
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
    details_url: Optional[str] = None,
) -> Dict[str, str]:
    results = parse_results(stream)
    selected = _selected_names(selected_regex)

    failed, flaky, passed, skipped, filtered = [], [], [], [], []
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

    ran = set(results)
    filtered = sorted(n for n in selected if n not in ran)

    counts = []
    if failed:
        counts.append(f"{len(failed)} failed")
    if flaky:
        counts.append(f"{len(flaky)} flaky")
    if passed:
        counts.append(f"{len(passed)} passed")
    if skipped:
        counts.append(f"{len(skipped)} skipped")
    if filtered:
        counts.append(f"{len(filtered)} not run")
    title = ", ".join(counts) if counts else "No tests ran"

    lines = []
    if selected:
        lines.append(f"{len(selected)} test(s) selected by the changed files in this PR.")
        lines.append("")
    lines.append("| Test | Result |")
    lines.append("| --- | --- |")
    for name in failed:
        lines.append(f"| `{name}` | :x: failed |")
    for name in flaky:
        lines.append(f"| `{name}` | :warning: flaky — failed, passed on re-run |")
    for name in filtered:
        lines.append(f"| `{name}` | :fast_forward: not run — in `flaky_tests.yaml` |")
    for name in skipped:
        lines.append(f"| `{name}` | :heavy_minus_sign: skipped |")
    for name in passed:
        lines.append(f"| `{name}` | :white_check_mark: passed |")

    if flaky:
        lines.append("")
        lines.append(
            "Flaky tests do not fail this check: `gotestsum --rerun-fails` reruns "
            "them and exits on the final result."
        )
    if details_url:
        lines.append("")
        lines.append(f"[Full logs]({details_url})")

    return {"title": title, "summary": "\n".join(lines)}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--jsonfile", required=True)
    parser.add_argument("--selected-regex", default=None)
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
            details_url=args.details_url,
        ),
        sys.stdout,
    )
    print()


if __name__ == "__main__":
    main()
