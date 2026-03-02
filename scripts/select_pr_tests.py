#!/usr/bin/env python3
"""
Reads a list of changed files (one per line, from stdin) and outputs a
-run-compatible regex of test functions to execute for PR integration tests.

Usage:
    git diff --name-only origin/master...HEAD | python3 scripts/select_pr_tests.py
    echo "datadog/resource_datadog_monitor.go" | python3 scripts/select_pr_tests.py

Exit behavior:
    - Outputs pipe-separated test function regex to stdout
    - Empty output means no relevant tests found (caller should skip tests)
"""

import os
import re
import sys
from typing import List

_RESOURCE_PATH_RE = re.compile(
    r"^datadog/(?:fwprovider/)?(resource_datadog_|data_source_datadog_)(.+)\.go$"
)
_TEST_PATH_RE = re.compile(r"^datadog/tests/.*_test\.go$")
_FUNC_RE = re.compile(r"^func (Test[A-Za-z0-9_]+)\(", re.MULTILINE)


def select_pr_tests(
    changed_files: List[str],
    repo_root: str,
    *,
    escape_for_make: bool = False,
) -> str:
    if not changed_files:
        return ""

    tests_dir = os.path.join(repo_root, "datadog", "tests")

    # Phase 1: Categorize changed files
    resource_types: set[str] = set()
    direct_test_files: set[str] = set()

    for f in changed_files:
        f = f.strip()
        if not f:
            continue

        if _TEST_PATH_RE.match(f):
            direct_test_files.add(os.path.join(repo_root, f))
            continue

        m = _RESOURCE_PATH_RE.match(f)
        if m:
            prefix = m.group(1)  # "resource_datadog_" or "data_source_datadog_"
            suffix = m.group(2)
            # Reconstruct the full type name: e.g. "datadog_monitor"
            type_name = prefix.replace("resource_", "").replace("data_source_", "") + suffix
            resource_types.add(type_name)

    # Phase 2: Find test files matching resource types via content search
    matched_test_files: set[str] = set()

    if resource_types:
        try:
            test_files = [
                os.path.join(tests_dir, fname)
                for fname in os.listdir(tests_dir)
                if fname.endswith("_test.go")
            ]
        except FileNotFoundError:
            test_files = []

        for rtype in resource_types:
            pattern = re.compile(
                rf'(?:resource|data)\s+"{re.escape(rtype)}"'
                rf'|"{re.escape(rtype)}\.'
            )
            for tf in test_files:
                try:
                    with open(tf) as fh:
                        if pattern.search(fh.read()):
                            matched_test_files.add(tf)
                except OSError:
                    continue

    # Phase 3: Combine
    all_test_files = matched_test_files | direct_test_files
    if not all_test_files:
        return ""

    # Phase 4: Extract test function names
    test_funcs: set[str] = set()
    for tf in all_test_files:
        try:
            with open(tf) as fh:
                for m in _FUNC_RE.finditer(fh.read()):
                    test_funcs.add(m.group(1))
        except OSError:
            continue

    if not test_funcs:
        return ""

    anchor = "$$" if escape_for_make else "$"
    return "|".join(f"^{name}{anchor}" for name in sorted(test_funcs))


def main() -> None:
    escape = "--escape-for-make" in sys.argv

    repo_root = os.path.join(os.path.dirname(__file__), "..")
    repo_root = os.path.abspath(repo_root)

    changed_files = [line for line in sys.stdin.read().splitlines() if line.strip()]

    result = select_pr_tests(changed_files, repo_root, escape_for_make=escape)
    if result:
        print(result)


if __name__ == "__main__":
    main()
