#!/usr/bin/env python3

import argparse
import re
import sys
from pathlib import Path

CASSETTE_EXTS = (".yaml", ".freeze")


def discover_tests(tests_dir: Path) -> set[str]:
    tests: set[str] = set()
    func_re = re.compile(r"^func (Test[A-Za-z0-9_]+)\(", re.MULTILINE)

    for path in tests_dir.glob("*_test.go"):
        tests.update(func_re.findall(path.read_text()))
    return tests


def discover_cassette_basenames(cassettes_dir: Path) -> set[str]:
    return {p.stem for p in cassettes_dir.iterdir() if p.suffix in CASSETTE_EXTS}


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--delete",
        action="store_true",
        help="Delete orphaned .yaml and .freeze files instead of just reporting them.",
    )
    args = parser.parse_args()

    repo_root = Path(__file__).resolve().parent.parent
    tests_dir = repo_root / "datadog" / "tests"
    cassettes_dir = tests_dir / "cassettes"

    tests = discover_tests(tests_dir)
    orphans = sorted(discover_cassette_basenames(cassettes_dir) - tests)

    if not orphans:
        return 0

    status = "deleted" if args.delete else "orphaned"
    for base in orphans:
        for ext in CASSETTE_EXTS:
            path = cassettes_dir / f"{base}{ext}"
            if not path.exists():
                continue
            print(f"{status}: {path.relative_to(repo_root)}")
            if args.delete:
                path.unlink()

    if args.delete:
        return 0

    return 1


if __name__ == "__main__":
    sys.exit(main())
