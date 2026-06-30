#!/usr/bin/env bash
# Throwaway data-source coverage analyzer — NOT maintained.
# Runs the tool (which reuses the real generator internals) against a spec.
#
# Usage:
#   ./run.sh [SPEC] [-- extra flags]
#
# SPEC defaults to the upstream Datadog v2 spec downloaded to /tmp; override with
# a path, e.g. ./run.sh .generator/V2/openapi.yaml
#
# Examples:
#   ./run.sh                                  # default spec, summary only
#   ./run.sh /tmp/dd_openapi_v2.yaml
#   ./run.sh /tmp/dd_openapi_v2.yaml -- --list-fails
#   ./run.sh /tmp/dd_openapi_v2.yaml -- --json /tmp/coverage.json
set -euo pipefail

# Resolve repo + module dirs relative to this script.
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"   # .generator-v2/tools/coverage
GEN_DIR="$(cd "$HERE/../.." && pwd)"                    # .generator-v2
REPO_DIR="$(cd "$GEN_DIR/.." && pwd)"                   # repo root

SPEC_DEFAULT="/tmp/dd_openapi_v2.yaml"
SPEC_URL="https://raw.githubusercontent.com/DataDog/datadog-api-client-go/refs/heads/master/.generator/schemas/v2/openapi.yaml"

# First non-flag arg is the spec path; everything after `--` is passed through.
SPEC="$SPEC_DEFAULT"
PASS_ARGS=()
if [[ $# -gt 0 && "$1" != "--" ]]; then SPEC="$1"; shift; fi
if [[ "${1:-}" == "--" ]]; then shift; PASS_ARGS=("$@"); fi

# Auto-fetch the upstream spec to /tmp if using the default and it's missing.
if [[ "$SPEC" == "$SPEC_DEFAULT" && ! -f "$SPEC" ]]; then
  echo "fetching upstream v2 spec -> $SPEC" >&2
  curl -sL "$SPEC_URL" -o "$SPEC"
fi

# Run from the module dir so go resolves the local module + internal packages.
cd "$GEN_DIR"
go run ./tools/coverage --spec "$SPEC" "${PASS_ARGS[@]}"
