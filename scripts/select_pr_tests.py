#!/usr/bin/env python3
"""
Reads a list of changed files (one per line, from stdin) and outputs a
-run-compatible regex of test functions to execute for PR integration tests.

Usage:
    git diff --name-only origin/master...HEAD | python3 scripts/select_pr_tests.py
    echo "datadog/resource_datadog_monitor.go" | python3 scripts/select_pr_tests.py

    # Read test-file contents from a git ref instead of the working tree. Used
    # by the pull_request_target workflow, which checks out the trusted base but
    # needs the PR head's version of the tests to discover newly added ones:
    git diff ... | python3 scripts/select_pr_tests.py --git-ref <pr-head-sha>

Exit behavior:
    - Outputs pipe-separated test function regex to stdout
    - Empty output means no relevant tests found (caller should skip tests)
"""

import os
import re
import subprocess
import sys
from typing import List, Optional

_RESOURCE_PATH_RE = re.compile(
    r"^datadog/(?:fwprovider/)?(resource_datadog_|data_source_datadog_)(.+)\.go$"
)
_TEST_PATH_RE = re.compile(r"^datadog/tests/.*_test\.go$")
_FUNC_RE = re.compile(r"^func (Test[A-Za-z0-9_]+)\(", re.MULTILINE)


def _git_list_test_files(repo_root: str, ref: str) -> List[str]:
    """List datadog/tests/*_test.go paths present at the given ref."""
    try:
        out = subprocess.run(
            ["git", "-C", repo_root, "ls-tree", "-r", "--name-only", ref, "datadog/tests"],
            check=True,
            capture_output=True,
            text=True,
        ).stdout
    except (subprocess.CalledProcessError, OSError):
        return []
    return [p for p in out.splitlines() if p.endswith("_test.go")]


def _git_read(repo_root: str, ref: str, relpath: str) -> Optional[str]:
    """Return the contents of relpath at the given ref, or None if absent.

    Only reads blob text (never executes it), so it is safe to point at an
    untrusted PR head.
    """
    try:
        return subprocess.run(
            ["git", "-C", repo_root, "show", f"{ref}:{relpath}"],
            check=True,
            capture_output=True,
        ).stdout.decode("utf-8", errors="replace")
    except (subprocess.CalledProcessError, OSError):
        return None


def _fs_list_test_files(repo_root: str) -> List[str]:
    """List datadog/tests/*_test.go paths in the working tree."""
    tests_dir = os.path.join(repo_root, "datadog", "tests")
    try:
        return [
            os.path.join("datadog", "tests", fname)
            for fname in os.listdir(tests_dir)
            if fname.endswith("_test.go")
        ]
    except FileNotFoundError:
        return []


def _fs_read(repo_root: str, relpath: str) -> Optional[str]:
    """Return the contents of relpath from the working tree, or None if absent."""
    try:
        with open(os.path.join(repo_root, relpath)) as fh:
            return fh.read()
    except OSError:
        return None


def select_pr_tests(
    changed_files: List[str],
    repo_root: str,
    *,
    escape_for_make: bool = False,
    git_ref: Optional[str] = None,
) -> str:
    if not changed_files:
        return ""

    # Resolve test files and their contents either from a git ref (the PR head,
    # so newly added tests are visible) or from the working tree (local use).
    if git_ref:
        def list_test_files() -> List[str]:
            return _git_list_test_files(repo_root, git_ref)

        def read_content(relpath: str) -> Optional[str]:
            return _git_read(repo_root, git_ref, relpath)
    else:
        def list_test_files() -> List[str]:
            return _fs_list_test_files(repo_root)

        def read_content(relpath: str) -> Optional[str]:
            return _fs_read(repo_root, relpath)

    # Phase 1: Categorize changed files (repo-relative paths throughout)
    resource_types: set[str] = set()
    direct_test_files: set[str] = set()

    for f in changed_files:
        f = f.strip()
        if not f:
            continue

        if _TEST_PATH_RE.match(f):
            direct_test_files.add(f)
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
        patterns = [
            re.compile(
                rf'(?:resource|data)\s+"{re.escape(rtype)}"'
                rf'|"{re.escape(rtype)}\.'
            )
            for rtype in resource_types
        ]
        for tf in list_test_files():
            content = read_content(tf)
            if content is None:
                continue
            if any(p.search(content) for p in patterns):
                matched_test_files.add(tf)

    # Phase 3: Combine
    all_test_files = matched_test_files | direct_test_files
    if not all_test_files:
        return ""

    # Phase 4: Extract test function names
    test_funcs: set[str] = set()
    for tf in all_test_files:
        content = read_content(tf)
        if content is None:
            continue
        for m in _FUNC_RE.finditer(content):
            test_funcs.add(m.group(1))

    if not test_funcs:
        return ""

    anchor = "$$" if escape_for_make else "$"
    return "|".join(f"^{name}{anchor}" for name in sorted(test_funcs))


def main() -> None:
    escape = "--escape-for-make" in sys.argv

    git_ref = None
    if "--git-ref" in sys.argv:
        idx = sys.argv.index("--git-ref")
        if idx + 1 < len(sys.argv):
            git_ref = sys.argv[idx + 1]

    repo_root = os.path.join(os.path.dirname(__file__), "..")
    repo_root = os.path.abspath(repo_root)

    changed_files = [line for line in sys.stdin.read().splitlines() if line.strip()]

    result = select_pr_tests(
        changed_files, repo_root, escape_for_make=escape, git_ref=git_ref
    )
    if result:
        print(result)


if __name__ == "__main__":
    main()
