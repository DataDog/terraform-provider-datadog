#!/usr/bin/env bash
#
# Reads flaky_tests.yaml and outputs a -skip-compatible regex
# (pipe-separated test names). Exits with empty output if no tests to skip.
#
set -euo pipefail

SKIP_FILE="${1:-flaky_tests.yaml}"

if [ ! -f "$SKIP_FILE" ]; then
  exit 0
fi

# Extract test names from YAML lines like "  - test: TestAccFoo_Bar"
# Uses sed instead of grep -P for macOS/Linux portability
TESTS=$(sed -n 's/^[[:space:]]*-\{0,1\}[[:space:]]*test:[[:space:]]*//p' "$SKIP_FILE" | sed 's/[[:space:]]*$//' | grep -v '^$' || true)

if [ -z "$TESTS" ]; then
  exit 0
fi

# Output as pipe-separated regex
echo "$TESTS" | paste -sd '|' -
