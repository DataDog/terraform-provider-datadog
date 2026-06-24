#!/usr/bin/env bash
# Local dev helper (gitignored): generate one mini-OAS data source and check the output is clean.
#
#   ./tfgen-gen.sh <slice>             generate <slice> into $OUT, verify clean
#   ./tfgen-gen.sh <slice> --replace   also overwrite the live provider data source
#
# Builds tfgen, annotates the mini slices, generates <slice>, then verifies the
# run succeeded and the emitted Go is gofmt-clean. With --replace it overwrites
# datadog/fwprovider/data_source_datadog_<slice>.go in place, renaming the
# generated constructor to the one the provider registry already references so
# the tree still compiles. Then build/test as normal: make build && make test.
set -euo pipefail

SLICE="${1:-datastore}"
MODE="${2:-}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MINI_OAS="$ROOT/.generator-v2/internal/testdata/mini-oas"
GEN_TEST="$MINI_OAS/scripts/gen-test"
OUT="${OUT:-$ROOT/.tfgen-out}"   # in-repo, gitignored; override with OUT=...
PROVIDER_DIR="$ROOT/datadog/fwprovider"

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

# Gate: every check below must pass before we touch the live provider (--replace).
echo
GEN="$OUT/data_source_datadog_$SLICE.go"

# 1. The generator's own report flagged no failed artifacts (exit code).
if [[ "$gen_ok" -ne 1 ]]; then
  echo "FAIL: generate reported a failed artifact (see report above)"
  exit 1
fi

# 2. The artifact was actually emitted (a skipped/unsupported kind exits 0 but writes nothing).
if [[ ! -f "$GEN" ]]; then
  echo "FAIL: no file generated for '$SLICE' (artifact skipped — see report above)"
  exit 1
fi

# 3. The emitted Go is gofmt-clean.
unclean="$(gofmt -l "$OUT"/*.go 2>/dev/null || true)"
if [[ -n "$unclean" ]]; then
  echo "FAIL: not gofmt-clean:"
  echo "$unclean"
  exit 1
fi
echo "PASS: '$SLICE' generated and gofmt-clean"

if [[ "$MODE" != "--replace" ]]; then
  ls "$OUT"
  exit 0
fi

# --- --replace: overwrite the live provider data source -----------------------
LIVE="$PROVIDER_DIR/data_source_datadog_$SLICE.go"
if [[ ! -f "$LIVE" ]]; then
  echo "ERROR: no framework data source to replace at $LIVE" >&2
  echo "       (--replace only handles data sources that already live in datadog/fwprovider/)" >&2
  exit 1
fi

# The constructor is the only symbol referenced outside the file (the provider
# registry in framework_provider.go). Keep the live name so the tree compiles.
live_ctor="$(grep -oE 'New[A-Za-z0-9]+DataSource' "$LIVE" | head -1)"
gen_ctor="$(grep -oE 'New[A-Za-z0-9]+DataSource' "$GEN" | head -1)"
if [[ -z "$live_ctor" || -z "$gen_ctor" ]]; then
  echo "ERROR: could not find a New...DataSource constructor in live or generated file" >&2
  exit 1
fi

echo "==> overwriting $LIVE (constructor $gen_ctor -> $live_ctor)"
cp "$GEN" "$LIVE"
perl -pi -e "s/\b\Q$gen_ctor\E\b/$live_ctor/g" "$LIVE"
gofmt -w "$LIVE"

echo "PASS: replaced '$SLICE' in the provider. Now build/test as normal:"
echo "        make build && make test"
