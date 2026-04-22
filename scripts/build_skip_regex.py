#!/usr/bin/env python3
"""
Reads flaky_tests.yaml and outputs a -skip-compatible regex
(pipe-separated, anchored test names). Empty output means no tests to skip.

Usage:
    python3 scripts/build_skip_regex.py [--escape-for-make] [path/to/flaky_tests.yaml]
"""

import os
import re
import sys


def build_skip_regex(path: str, *, escape_for_make: bool = False) -> str:
    if not os.path.isfile(path):
        return ""

    tests: list[str] = []
    with open(path) as f:
        for line in f:
            m = re.match(r"^\s*-?\s*test:\s*(.+)", line)
            if m:
                name = m.group(1).strip()
                if name:
                    tests.append(name)

    if not tests:
        return ""

    anchor = "$$" if escape_for_make else "$"
    return "|".join(f"^{t}{anchor}" for t in tests)


def main() -> None:
    escape = "--escape-for-make" in sys.argv
    args = [a for a in sys.argv[1:] if a != "--escape-for-make"]
    path = args[0] if args else "flaky_tests.yaml"

    result = build_skip_regex(path, escape_for_make=escape)
    if result:
        print(result)


if __name__ == "__main__":
    main()
