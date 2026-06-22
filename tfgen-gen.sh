#!/usr/bin/env bash
# Local dev helper (gitignored): generate one mini-OAS data source and check the output is clean.
#
#   ./tfgen-gen.sh <slice>     e.g. ./tfgen-gen.sh datastore
#
# Builds tfgen, annotates the mini slices, generates <slice> into $OUT, then
# verifies the run succeeded and the emitted Go is gofmt-clean.
set -euo pipefail

SLICE="${1:-datastore}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MINI_OAS="$ROOT/.generator-v2/internal/testdata/mini-oas"
GEN_TEST="$MINI_OAS/scripts/gen-test"
OUT="${OUT:-$ROOT/.tfgen-out}"   # in-repo, gitignored; override with OUT=...

cd "$ROOT"

echo "==> building tfgen"
make tfgen-build >/dev/null

echo "==> annotating mini slices"
( cd "$MINI_OAS/scripts" && python3 _annotate.py >/dev/null )

SPEC="$GEN_TEST/$SLICE.yaml"
if [[ ! -f "$SPEC" ]]; then
  echo "ERROR: no annotated slice '$SLICE' at $SPEC" >&2
  echo "       add '$SLICE' to SAMPLE in _annotate.py (or hand-annotate the slice)." >&2
  echo "       available:" >&2
  ( cd "$GEN_TEST" && ls *.yaml 2>/dev/null | sed 's/\.yaml$//; s/^/         - /' ) >&2
  exit 1
fi

echo "==> generating '$SLICE' -> $OUT"
rm -rf "$OUT"
gen_ok=1
./bin/tfgen generate --spec "$SPEC" --output-root "$OUT" --report - || gen_ok=0

echo
if [[ "$gen_ok" -ne 1 ]]; then
  echo "FAIL: generate reported a failed artifact (see report above)"
  exit 1
fi

unclean="$(gofmt -l "$OUT"/*.go 2>/dev/null || true)"
if [[ -n "$unclean" ]]; then
  echo "FAIL: not gofmt-clean:"
  echo "$unclean"
  exit 1
fi

echo "PASS: '$SLICE' generated and gofmt-clean"
ls "$OUT"
