#!/usr/bin/env python3
"""
Assigns tests to shards for parallel CI execution.

Scans datadog/tests/*_test.go for test function declarations, assigns each
to a shard via hash(test_name) % total_shards, and outputs a -run-compatible
regex for the requested shard.

Usage:
    python3 scripts/shard_tests.py --shard 0 --total 4
    python3 scripts/shard_tests.py --shard 0 --total 4 --escape-for-make
"""

import hashlib
import os
import re
import sys

_FUNC_RE = re.compile(r"^func (Test[A-Za-z0-9_]+)\(", re.MULTILINE)


def shard_tests(
    shard_index: int,
    total_shards: int,
    repo_root: str,
    *,
    escape_for_make: bool = False,
) -> str:
    tests_dir = os.path.join(repo_root, "datadog", "tests")

    all_tests: list[str] = []
    for fname in sorted(os.listdir(tests_dir)):
        if not fname.endswith("_test.go"):
            continue
        with open(os.path.join(tests_dir, fname)) as fh:
            for m in _FUNC_RE.finditer(fh.read()):
                all_tests.append(m.group(1))

    shard: list[str] = []
    for test in sorted(all_tests):
        h = int(hashlib.md5(test.encode()).hexdigest(), 16)
        if h % total_shards == shard_index:
            shard.append(test)

    if not shard:
        return ""

    anchor = "$$" if escape_for_make else "$"
    return "|".join(f"^{name}{anchor}" for name in shard)


def main() -> None:
    import argparse

    parser = argparse.ArgumentParser(description="Assign tests to CI shards")
    parser.add_argument("--shard", type=int, required=True, help="Shard index (0-based)")
    parser.add_argument("--total", type=int, required=True, help="Total number of shards")
    parser.add_argument(
        "--escape-for-make",
        action="store_true",
        help="Use $$ instead of $ for Make compatibility",
    )
    args = parser.parse_args()

    if args.shard < 0 or args.shard >= args.total:
        print(f"Error: --shard must be in [0, {args.total})", file=sys.stderr)
        sys.exit(1)

    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    result = shard_tests(args.shard, args.total, repo_root, escape_for_make=args.escape_for_make)
    if result:
        print(result)


if __name__ == "__main__":
    main()
