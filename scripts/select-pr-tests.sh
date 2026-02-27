#!/usr/bin/env bash
#
# Reads a list of changed files (one per line, from stdin) and outputs a
# -run-compatible regex of test functions to execute for PR integration tests.
#
# Usage:
#   git diff --name-only origin/master...HEAD | scripts/select-pr-tests.sh
#   echo "datadog/resource_datadog_monitor.go" | scripts/select-pr-tests.sh
#
# Exit behavior:
#   - Outputs pipe-separated test function regex to stdout
#   - Empty output means no relevant tests found (caller should skip tests)
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TESTS_DIR="${REPO_ROOT}/datadog/tests"

# Collect changed files from stdin
CHANGED_FILES=""
while IFS= read -r line; do
  [ -z "$line" ] && continue
  CHANGED_FILES="${CHANGED_FILES}${line}"$'\n'
done

if [ -z "$CHANGED_FILES" ]; then
  exit 0
fi

# ---- Phase 1: Categorize changed files and extract resource type names ----

RESOURCE_TYPES=""
DIRECT_TEST_FILES=""

while IFS= read -r file; do
  [ -z "$file" ] && continue

  basename_file="$(basename "$file")"

  # Directly changed test files in datadog/tests/
  if echo "$file" | grep -qE '^datadog/tests/.*_test\.go$'; then
    DIRECT_TEST_FILES="${DIRECT_TEST_FILES}${REPO_ROOT}/${file}"$'\n'
    continue
  fi

  # Resource/datasource files (SDKv2: datadog/, Framework: datadog/fwprovider/)
  if echo "$file" | grep -qE '^datadog/(fwprovider/)?(resource_datadog_|data_source_datadog_).*\.go$'; then
    # Extract resource type name: strip prefix and .go suffix
    type_name=$(echo "$basename_file" | sed -E 's/^(resource_|data_source_)//' | sed 's/\.go$//')
    if [ -n "$type_name" ]; then
      RESOURCE_TYPES="${RESOURCE_TYPES}${type_name}"$'\n'
    fi
  fi

  # All other files (shared code, docs, CI, etc.) are ignored for test selection
done <<< "$CHANGED_FILES"

# Deduplicate resource types
if [ -n "$RESOURCE_TYPES" ]; then
  RESOURCE_TYPES=$(echo "$RESOURCE_TYPES" | sort -u | grep -v '^$' || true)
fi

# ---- Phase 2: Find test files matching resource types via content search ----

MATCHED_TEST_FILES=""

if [ -n "$RESOURCE_TYPES" ]; then
  while IFS= read -r rtype; do
    [ -z "$rtype" ] && continue

    # Search test files for references to this resource type:
    #   - resource "datadog_monitor" (HCL resource blocks)
    #   - data "datadog_monitor" (HCL data source blocks)
    #   - "datadog_monitor. (resource references in test checks)
    matched=$(grep -rlE "(resource|data) \"${rtype}\"|\"${rtype}\." "${TESTS_DIR}/"*_test.go 2>/dev/null || true)
    if [ -n "$matched" ]; then
      MATCHED_TEST_FILES="${MATCHED_TEST_FILES}${matched}"$'\n'
    fi
  done <<< "$RESOURCE_TYPES"
fi

# ---- Phase 3: Combine direct test files and content-matched test files ----

ALL_TEST_FILES=""
if [ -n "$MATCHED_TEST_FILES" ]; then
  ALL_TEST_FILES="${MATCHED_TEST_FILES}"
fi
if [ -n "$DIRECT_TEST_FILES" ]; then
  ALL_TEST_FILES="${ALL_TEST_FILES}${DIRECT_TEST_FILES}"
fi

if [ -z "$ALL_TEST_FILES" ]; then
  exit 0
fi

# Deduplicate
ALL_TEST_FILES=$(echo "$ALL_TEST_FILES" | sort -u | grep -v '^$' || true)

if [ -z "$ALL_TEST_FILES" ]; then
  exit 0
fi

# ---- Phase 4: Extract test function names from matched files ----

TEST_FUNCS=""
while IFS= read -r test_file; do
  [ -z "$test_file" ] && continue
  [ ! -f "$test_file" ] && continue

  # Extract "func TestFooBar(" -> "TestFooBar"
  funcs=$(grep -oE '^func (Test[A-Za-z0-9_]+)\(' "$test_file" | sed 's/^func //; s/($//' || true)
  if [ -n "$funcs" ]; then
    TEST_FUNCS="${TEST_FUNCS}${funcs}"$'\n'
  fi
done <<< "$ALL_TEST_FILES"

if [ -z "$TEST_FUNCS" ]; then
  exit 0
fi

# Deduplicate and output as pipe-separated regex
# Wrap each name with ^ and $ anchors to avoid partial matches, then join with |
echo "$TEST_FUNCS" | sort -u | grep -v '^$' | sed 's/.*/^&$/' | paste -sd '|' -
