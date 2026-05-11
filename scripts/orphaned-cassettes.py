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


def discover_cassettes(cassettes_dir: Path) -> dict[str, list[Path]]:
    cassettes: dict[str, list[Path]] = {}
    for ext in CASSETTE_EXTS:
        for path in cassettes_dir.rglob(f"*{ext}"):
            rel = path.relative_to(cassettes_dir)
            if len(rel.parts) < 2:
                continue
            name = path.stem if len(rel.parts) == 2 else rel.parts[1]
            cassettes.setdefault(name, []).append(path)
    return cassettes


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
    cassettes = discover_cassettes(cassettes_dir)
    orphans = sorted(set(cassettes) - tests)

    if not orphans:
        return 0

    status = "deleted" if args.delete else "orphaned"
    for name in orphans:
        for path in sorted(cassettes[name]):
            print(f"{status}: {path.relative_to(repo_root)}")
            if args.delete:
                path.unlink()

    if args.delete:
        return 0

    return 1


if __name__ == "__main__":
    sys.exit(main())
