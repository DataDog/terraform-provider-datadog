#!/usr/bin/env bash
# Headless spine for generating a Datadog data source and opening a draft PR.
#
# Deterministic orchestrator: validate -> slice -> generate -> gate -> whitelist
# -> docs -> build -> branch -> commit -> draft PR. The only non-deterministic
# step is a single, tightly-scoped LLM call for the runtime-risk narrative, which
# degrades gracefully (a failed scan never blocks the PR — it flags manual review).
#
# Safety model: every fork is an explicit flag or a hard failure. Nothing runs on
# a base branch, every PR is a draft, and the verification disclaimer is always
# present — the worst case is a draft PR a human closes.
#
# stdout carries ONLY the final JSON result. All human logs go to stderr.
# Exit 0 = success; nonzero = failure (with a JSON {status:"failed",...} on stdout).

set -euo pipefail

# This script's own directory, resolved before we cd elsewhere, so the prompt
# files under prompts/ still load. Copy them to a tmp dir OUTSIDE the working tree:
# we later `git checkout` the base branch, which resets the tree to that branch's
# content — if the base doesn't carry prompts/, they'd vanish mid-run.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROMPTS_DIR="$(mktemp -d -t tfgen-prompts.XXXXXX)"
cp "$SCRIPT_DIR/prompts/"*.md "$PROMPTS_DIR/" 2>/dev/null || true

# Remove this run's scratch files on any exit. REPORT is excluded (its path is
# returned in the result); only a spec we curled is removed, never a --spec file.
cleanup() {
  if [ -n "${PROMPTS_DIR:-}" ]; then rm -rf "$PROMPTS_DIR" 2>/dev/null || true; fi
  rm -f "${CLAUDE_COST_FILE:-}" "${RISK_PROMPT_FILE:-}" \
        "${PROSE_PROMPT_FILE:-}" "${PR_BODY_FILE:-}" 2>/dev/null || true
  if [ "${SPEC_IS_TEMP:-0}" = 1 ]; then rm -f "${SPEC:-}" 2>/dev/null || true; fi
  return 0
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Logging + structured failure
# ---------------------------------------------------------------------------
STAGE="init"
jlog() { printf '%s\n' "[$STAGE] $*" >&2; }

# emit_result <json> — the single exit point for a terminal result: print it to
# stdout and, when --output-json was given, also write it there. The trailing
# `return 0` keeps a skipped file-write from failing the caller under `set -e`.
emit_result() {
  printf '%s\n' "$1"
  [ -n "${OUTPUT_JSON:-}" ] && printf '%s\n' "$1" >"$OUTPUT_JSON"
  return 0
}

# die <message> — emit a failure result on stdout and exit nonzero.
die() {
  local msg="$1"
  jlog "FATAL: $msg"
  # If the branch was made but nothing was committed yet, delete it. If a commit
  # already exists, keep it so a push or PR can be retried.
  if [ "${BRANCH_CREATED:-0}" = 1 ] && [ "${COMMITTED:-0}" = 0 ]; then
    # Discard everything this run touched, then force-switch back: a bare checkout
    # would refuse while tracked files (e.g. framework_provider.go under
    # --overwrites) are still modified, stranding us on the throwaway branch.
    git reset --hard >/dev/null 2>&1 || true
    git clean -fdq datadog/fwprovider datadog/tests docs/data-sources >/dev/null 2>&1 || true
    git checkout -f "${ORIG_BRANCH:-master}" >/dev/null 2>&1 || true
    git branch -D "$BRANCH" >/dev/null 2>&1 || true
  fi
  emit_result "$(jq -n \
    --arg status failed --arg stage "$STAGE" --arg error "$msg" \
    --arg artifact_name "${ARTIFACT_NAME:-}" --arg branch "${BRANCH:-}" \
    --argjson metrics "$(metrics_json)" \
    '{status:$status, stage:$stage, error:$error, artifact_name:$artifact_name, branch:$branch, verified:false, metrics:$metrics}')"
  exit 1
}

usage() {
  cat >&2 <<'EOF'
Usage: generate_headless.sh --artifact-name NAME --cardinality {singular|plural} \
         (--read OP | --search OP) [flags]

Required:
  --artifact-name NAME     snake_case, no datadog_ prefix (^[a-z][a-z0-9_]*$, <=64)
  --cardinality VALUE      singular | plural  (explicit — never inferred)
  --read OP / --search OP  operationId(s); >=1 required. Plural: list op in --read.

Optional:
  --tf-description TEXT     doc string (default derived from the name)
  --overwrites CTOR        retire a hand-written constructor, e.g. NewDatadogTeamDataSource
  --service NAME           PR title [prefix] (default: derived from the op's spec tag)
  --spec PATH              full v2 OAS file (default: curl upstream)
  --spec-ref REF           git ref of datadog-api-client-go to curl (default: master)
  --base BRANCH            branch the PR targets and is built from (default: current branch)
  --branch NAME            feature branch (default: generate/datadog_<name>_datasource)
  --no-pr                  stop after commit; do not push or open a PR
  --output-json PATH       also write the final result JSON here
EOF
  exit 2
}

# Running tally of Claude usage, written to a file so it survives the subshells
# that command substitution runs call_claude in.
CLAUDE_COST_FILE=""

# call_claude <prompt-file> — print the model's JSON reply (fences stripped) and
# return 0, or return 1 if claude is missing, errors, or returns non-JSON. The
# whole file is passed as one argument, so any $ or backticks in it stay literal.
call_claude() {
  local pf="$1" raw result scratch to
  command -v claude >/dev/null 2>&1 || return 1
  # timeout isn't guaranteed (notably on macOS), so wrap only if it's present.
  to=""; for t in timeout gtimeout; do command -v "$t" >/dev/null 2>&1 && { to="$t 180"; break; }; done
  # Run from an empty scratch dir with --strict-mcp-config so the call ignores the
  # repo's project CLAUDE.md/settings and any auto-loaded MCP servers; the prompt
  # file is self-contained, so nothing project-specific is needed.
  scratch="$(mktemp -d -t tfgen-claude.XXXXXX)"
  raw="$( cd "$scratch" && $to claude -p "$(cat "$pf")" --strict-mcp-config --output-format json --max-turns 1 2>/dev/null || true )"
  rm -rf "$scratch" 2>/dev/null || true
  # Record what this call cost even if the reply is later unusable — we still paid.
  [ -n "$CLAUDE_COST_FILE" ] && printf '%s\n' "$raw" | jq -c \
    '{cost: (.total_cost_usd // .cost_usd // 0), in: ((.usage.input_tokens // 0) + (.usage.cache_read_input_tokens // 0) + (.usage.cache_creation_input_tokens // 0)), out: (.usage.output_tokens // 0)}' \
    >>"$CLAUDE_COST_FILE" 2>/dev/null || true
  result="$(printf '%s' "$raw" | jq -r 'if type=="object" then (.result // "") else . end' 2>/dev/null || true)"
  result="$(printf '%s' "$result" | sed -e 's/^```json//' -e 's/^```//' -e 's/```$//')"
  printf '%s' "$result" | jq -e . >/dev/null 2>&1 || return 1
  printf '%s' "$result"
}

# metrics_json — total runtime and Claude usage as one JSON object; safe anytime.
metrics_json() {
  local cost=0 intok=0 outtok=0 calls=0
  if [ -n "$CLAUDE_COST_FILE" ] && [ -s "$CLAUDE_COST_FILE" ]; then
    cost="$(jq -s '(map(.cost) | add // 0) | (. * 1000000 | round) / 1000000' "$CLAUDE_COST_FILE" 2>/dev/null || echo 0)"
    intok="$(jq -s 'map(.in) | add // 0' "$CLAUDE_COST_FILE" 2>/dev/null || echo 0)"
    outtok="$(jq -s 'map(.out) | add // 0' "$CLAUDE_COST_FILE" 2>/dev/null || echo 0)"
    calls="$(grep -c '' "$CLAUDE_COST_FILE" 2>/dev/null || echo 0)"
  fi
  jq -n --argjson rt "${SECONDS:-0}" --argjson cost "$cost" \
        --argjson intok "$intok" --argjson outtok "$outtok" --argjson calls "$calls" \
    '{runtime_seconds:$rt, claude_cost_usd:$cost, claude_input_tokens:$intok, claude_output_tokens:$outtok, claude_calls:$calls}'
}

# ---------------------------------------------------------------------------
# Args
# ---------------------------------------------------------------------------
ARTIFACT_NAME="" CARDINALITY="" READ_OP="" SEARCH_OP="" TF_DESCRIPTION=""
OVERWRITES="" SERVICE="" SPEC="" SPEC_REF="master" BASE="" BRANCH=""
NO_PR=0 OUTPUT_JSON=""

while [ $# -gt 0 ]; do
  case "$1" in
    --artifact-name) ARTIFACT_NAME="${2:-}"; shift 2 ;;
    --cardinality)   CARDINALITY="${2:-}"; shift 2 ;;
    --read)          READ_OP="${2:-}"; shift 2 ;;
    --search)        SEARCH_OP="${2:-}"; shift 2 ;;
    --tf-description) TF_DESCRIPTION="${2:-}"; shift 2 ;;
    --overwrites)    OVERWRITES="${2:-}"; shift 2 ;;
    --service)       SERVICE="${2:-}"; shift 2 ;;
    --spec)          SPEC="${2:-}"; shift 2 ;;
    --spec-ref)      SPEC_REF="${2:-}"; shift 2 ;;
    --base)          BASE="${2:-}"; shift 2 ;;
    --branch)        BRANCH="${2:-}"; shift 2 ;;
    --no-pr)         NO_PR=1; shift ;;
    --output-json)   OUTPUT_JSON="${2:-}"; shift 2 ;;
    -h|--help)       usage ;;
    *) echo "unknown flag: $1" >&2; usage ;;
  esac
done

# ---------------------------------------------------------------------------
# Stage: preconditions — fail fast before touching anything
# ---------------------------------------------------------------------------
STAGE="preconditions"

[ -n "$ARTIFACT_NAME" ] || die "missing --artifact-name"
[ -n "$CARDINALITY" ] || die "missing --cardinality (must be explicit: singular|plural)"
case "$CARDINALITY" in singular|plural) ;; *) die "invalid --cardinality '$CARDINALITY'" ;; esac
[ -n "$READ_OP" ] || [ -n "$SEARCH_OP" ] || die "need at least one of --read / --search"
if [ "$CARDINALITY" = plural ] && [ -z "$READ_OP" ]; then
  die "plural cardinality requires the collection GET in --read (not --search)"
fi
[[ "$ARTIFACT_NAME" =~ ^[a-z][a-z0-9_]*$ ]] || die "invalid --artifact-name '$ARTIFACT_NAME' (^[a-z][a-z0-9_]*$)"
[ "${#ARTIFACT_NAME}" -le 64 ] || die "--artifact-name too long (>64)"
# A name ending in _test makes the generated data_source_datadog_<name>.go a
# *_test.go file, which Go excludes from the build — a guaranteed compile failure.
[[ "$ARTIFACT_NAME" == *_test ]] && die "--artifact-name must not end in '_test' (Go would treat the generated file as a test file and drop it from the build)"

for tool in git gh python3 make curl jq; do
  command -v "$tool" >/dev/null 2>&1 || die "required tool not found: $tool"
done
gh auth status >/dev/null 2>&1 || die "gh is not authenticated (run: gh auth login)"
python3 -c 'import yaml' 2>/dev/null || die "python3 PyYAML not installed (pip install pyyaml)"

# Repo root = the dir holding the generator; search upward so this works from a subdir.
find_root() {
  local d; d="$(pwd)"
  while [ "$d" != "/" ]; do
    [ -f "$d/.generator-v2/cmd/tfgen/main.go" ] && { printf '%s\n' "$d"; return 0; }
    d="$(dirname "$d")"
  done
  return 1
}
ROOT="$(find_root)" || die "not inside a terraform-provider-datadog checkout (no .generator-v2/cmd/tfgen)"
cd "$ROOT"
jlog "repo root: $ROOT"
SLICER=".generator-v2/internal/testdata/mini-oas/scripts/slice_and_annotate.py"
[ -f "$SLICER" ] || die "slicer not found at $SLICER"

# Stop if there are uncommitted changes, so the commit only holds generated files.
[ -z "$(git status --porcelain)" ] || die "working tree is dirty; commit or stash first"

# Remember the starting branch so a failure can put us back on it.
ORIG_BRANCH="$(git rev-parse --abbrev-ref HEAD)"

# Default the base (PR target + branch-point) to the branch we're on, so headless
# generation cuts and merges into the current branch rather than master.
if [ -z "$BASE" ]; then
  [ "$ORIG_BRANCH" != "HEAD" ] || die "detached HEAD — pass --base explicitly (cannot default base to current branch)"
  BASE="$ORIG_BRANCH"
fi

# If we'll open a PR, the base must exist on origin — GitHub can't target a branch
# it doesn't have. Check now, before any work, so a local-only base fails fast
# instead of after the commit + push.
if [ "$NO_PR" -eq 0 ]; then
  git ls-remote --exit-code --heads origin "$BASE" >/dev/null 2>&1 \
    || die "base branch '$BASE' does not exist on origin; push it first, pass --base to an existing remote branch, or use --no-pr"
fi

BRANCH_CREATED=0
COMMITTED=0

# Branch name + collision check.
BRANCH="${BRANCH:-generate/datadog_${ARTIFACT_NAME}_datasource}"
if git rev-parse --verify --quiet "refs/heads/$BRANCH" >/dev/null; then
  die "branch already exists locally: $BRANCH"
fi
if git ls-remote --exit-code --heads origin "$BRANCH" >/dev/null 2>&1; then
  die "branch already exists on origin: $BRANCH"
fi

# Overwrite safety: an explicit --overwrites is intent (it lands on a branch, in a
# draft PR — nothing prod is touched). But if a hand-written source exists and the
# caller did NOT opt in, additive generation would double-register — so stop.
HANDWRITTEN="datadog/fwprovider/data_source_datadog_${ARTIFACT_NAME}.go"
if [ -z "$OVERWRITES" ] && [ -f "$HANDWRITTEN" ]; then
  die "a hand-written $HANDWRITTEN exists; pass --overwrites <Ctor> to retire it, or choose another --artifact-name"
fi

# Start the new branch from the branch the PR will target, so the PR shows only
# the generated files. Use origin's copy when it's reachable, else the local one.
BASE_REF=""
if git fetch --quiet origin "$BASE" 2>/dev/null; then
  BASE_REF="origin/$BASE"
elif git rev-parse --verify --quiet "refs/heads/$BASE" >/dev/null; then
  BASE_REF="$BASE"
else
  die "base branch '$BASE' not found on origin or locally"
fi
jlog "starting $BRANCH from $BASE_REF (PR will target $BASE)"
git checkout -b "$BRANCH" "$BASE_REF" >&2 || die "failed to create $BRANCH from $BASE_REF"
BRANCH_CREATED=1

# ---------------------------------------------------------------------------
# Stage: spec — resolve content (float to latest) but record provenance
# ---------------------------------------------------------------------------
STAGE="spec"
SPEC_SOURCE="" SPEC_SHA=""
if [ -n "$SPEC" ]; then
  [ -f "$SPEC" ] || die "--spec not found: $SPEC"
  SPEC_SOURCE="file:$SPEC"
  jlog "using provided spec: $SPEC"
else
  SPEC="$(mktemp -t tfgen-oas.XXXXXX.yaml)"
  SPEC_IS_TEMP=1
  local_url="https://raw.githubusercontent.com/DataDog/datadog-api-client-go/refs/heads/${SPEC_REF}/.generator/schemas/v2/openapi.yaml"
  jlog "curling upstream v2 spec ($SPEC_REF)"
  curl -fsSL "$local_url" -o "$SPEC" || die "failed to curl upstream spec from $local_url"
  [ -s "$SPEC" ] || die "curled spec is empty"
  SPEC_SOURCE="upstream:$SPEC_REF"
  # Provenance: resolve the ref -> commit SHA so 'latest' is auditable after the fact.
  SPEC_SHA="$(curl -fsSL "https://api.github.com/repos/DataDog/datadog-api-client-go/commits/${SPEC_REF}" 2>/dev/null | jq -r '.sha // empty' || true)"
  jlog "resolved ${SPEC_REF} -> ${SPEC_SHA:-<unresolved>}"
fi

# ---------------------------------------------------------------------------
# Stage: slice — annotate the chosen op(s) into a standalone OAS
# ---------------------------------------------------------------------------
STAGE="slice"
if [ -z "$TF_DESCRIPTION" ]; then
  thing="${ARTIFACT_NAME//_/ }"
  if [ "$CARDINALITY" = plural ]; then
    TF_DESCRIPTION="Use this data source to retrieve information about existing Datadog ${thing}."
  else
    TF_DESCRIPTION="Use this data source to retrieve information about an existing Datadog ${thing}."
  fi
fi

slice_args=(--spec "$SPEC" --artifact-name "$ARTIFACT_NAME"
            --cardinality "$CARDINALITY" --tf-description "$TF_DESCRIPTION")
[ -n "$READ_OP" ]   && slice_args+=(--read "$READ_OP")
[ -n "$SEARCH_OP" ] && slice_args+=(--search "$SEARCH_OP")
[ -n "$OVERWRITES" ] && slice_args+=(--overwrites "$OVERWRITES")

SLICE="$(python3 "$SLICER" "${slice_args[@]}")" || die "slice_and_annotate.py failed (see stderr above)"
[ -f "$SLICE" ] || die "slicer reported no output slice"
jlog "slice: $SLICE"

# Derive the service tag for the CI-enforced [prefix] if not supplied.
if [ -z "$SERVICE" ]; then
  SERVICE="$(python3 - "$SLICE" <<'PY' || true
import sys, yaml
spec = yaml.safe_load(open(sys.argv[1]))
for item in spec.get("paths", {}).values():
    if not isinstance(item, dict):
        continue
    for node in item.values():
        if isinstance(node, dict) and "x-datadog-tf-generator" in node:
            tags = node.get("tags") or []
            if tags:
                print(tags[0].lower().replace(" ", "-"))
                sys.exit(0)
PY
)"
fi
[ -n "$SERVICE" ] || die "could not derive --service from the spec tag; pass --service explicitly"
jlog "service prefix: [$SERVICE]"

# ---------------------------------------------------------------------------
# Stage: generate — run tfgen, capture the RunReport
# ---------------------------------------------------------------------------
STAGE="generate"
[ -x bin/tfgen ] || { jlog "building tfgen"; make tfgen-build >&2 || die "make tfgen-build failed"; }
REPORT="$(mktemp -t tfgen-report.XXXXXX.json)"
if ! ./bin/tfgen generate --spec "$SLICE" --emit-tests --report "$REPORT" >&2; then
  jlog "tfgen exited nonzero"
fi
[ -s "$REPORT" ] || die "tfgen wrote no report"

# ---------------------------------------------------------------------------
# Stage: gate — stop before committing on any failure/error diagnostic
# ---------------------------------------------------------------------------
STAGE="gate"
FAILED="$(jq -r '.summary.failed // 0' "$REPORT")"
ERROR_DIAGS="$(jq -c '[.artifacts[]?.diagnostics[]? | select(.severity=="error")]' "$REPORT")"
if [ "$FAILED" != "0" ] || [ "$ERROR_DIAGS" != "[]" ]; then
  jlog "GATE FAILED — failed=$FAILED errors=$ERROR_DIAGS"
  # Discard partial output and drop the branch we just made.
  git checkout -- datadog/ 2>/dev/null || true
  git clean -fdq datadog/fwprovider datadog/tests docs/data-sources 2>/dev/null || true
  if [ "${BRANCH_CREATED:-0}" = 1 ]; then
    git checkout "${ORIG_BRANCH:-master}" >/dev/null 2>&1 || true
    git branch -D "$BRANCH" >/dev/null 2>&1 || true
  fi
  emit_result "$(jq -n --arg status failed --arg stage gate \
        --argjson failed "$FAILED" --argjson errors "$ERROR_DIAGS" \
        --arg artifact_name "$ARTIFACT_NAME" \
        --argjson metrics "$(metrics_json)" \
        '{status:$status, stage:$stage, error:"generation gate failed", failed:$failed, error_diagnostics:$errors, artifact_name:$artifact_name, verified:false, metrics:$metrics}')"
  exit 1
fi
# warning/info do not gate — carry them forward.
WARN_DIAGS="$(jq -c '[.artifacts[]?.diagnostics[]? | select(.severity!="error")]' "$REPORT")"
SPEC_HASH="$(jq -r '.spec_hash // empty' "$REPORT")"
jlog "gate passed (failed=0, no error diagnostics)"

# ---------------------------------------------------------------------------
# Stage: docs + build
# ---------------------------------------------------------------------------
STAGE="docs"
DOCS_FILE="docs/data-sources/${ARTIFACT_NAME}.md"
make docs >&2 || die "make docs failed"

# tfplugindocs regenerates the WHOLE docs/ tree, so any pre-existing drift (often
# just a Terraform-version rendering difference) surfaces as changed files that
# aren't ours. Restore everything under docs/ except our artifact's page, so the
# whitelist and commit see only the file we meant to add.
while IFS= read -r line; do
  [ -z "$line" ] && continue
  status="${line:0:2}"; path="${line:3}"
  path="${path#\"}"; path="${path%\"}"
  [ "$path" = "$DOCS_FILE" ] && continue
  if [ "$status" = "??" ]; then
    rm -f "$path" 2>/dev/null || true
  else
    git checkout -- "$path" >/dev/null 2>&1 || true
  fi
  jlog "reverted unrelated doc drift: $path"
done < <(git status --porcelain -- docs/)

[ -f "$DOCS_FILE" ] || die "make docs produced no $DOCS_FILE — registration did not take"

STAGE="build"
make build >&2 || die "make build failed (generated code does not compile)"
jlog "docs + build clean"

# ---------------------------------------------------------------------------
# Stage: whitelist — no human eyeballs git status, so assert it explicitly
# ---------------------------------------------------------------------------
STAGE="whitelist"
declare -a ALLOWED=(
  "datadog/fwprovider/data_source_datadog_${ARTIFACT_NAME}.go"
  "datadog/tests/data_source_datadog_${ARTIFACT_NAME}_test.go"
  "datadog/fwprovider/datasources_generated.go"
  "$DOCS_FILE"
)
[ -n "$OVERWRITES" ] && ALLOWED+=("datadog/fwprovider/framework_provider.go")

is_allowed() { local f="$1"; for a in "${ALLOWED[@]}"; do [ "$f" = "$a" ] && return 0; done; return 1; }

UNEXPECTED=()
while IFS= read -r line; do
  [ -z "$line" ] && continue
  path="${line:3}"           # strip the "XY " status prefix
  path="${path#\"}"; path="${path%\"}"
  is_allowed "$path" || UNEXPECTED+=("$path")
done < <(git status --porcelain)

if [ "${#UNEXPECTED[@]}" -gt 0 ]; then
  jlog "UNEXPECTED changed files: ${UNEXPECTED[*]}"
  die "files changed outside the generated-artifact whitelist: ${UNEXPECTED[*]}"
fi
jlog "changed-file whitelist clean"

CHANGED_JSON="$(git status --porcelain | sed 's/^...//' | jq -R . | jq -s .)"

# ---------------------------------------------------------------------------
# Stage: risk-scan — model call over prompts/risk-scan.md (advisory; degrades gracefully)
# ---------------------------------------------------------------------------
STAGE="risk-scan"
CLAUDE_COST_FILE="$(mktemp -t tfgen-claude-cost.XXXXXX.jsonl)"

# Deterministic flags first — a grep is more reliable than a model for these.
MECH_BULLETS=()
grep -qiE 'x-pagination|"?paginated"?[[:space:]]*:' "$SLICE" 2>/dev/null && \
  MECH_BULLETS+=("Endpoint appears paginated — confirm the generated read uses the SDK's ...WithPagination method, or results may be truncated.")
grep -qE '^[[:space:]]*enum:' "$SLICE" 2>/dev/null && \
  MECH_BULLETS+=("Schema contains enums — a new/unknown enum value can trip the SDK silent-empty trap; require cassette verification.")
grep -qE 'format:[[:space:]]*int64' "$SLICE" 2>/dev/null && \
  MECH_BULLETS+=("Wide (int64) integers present — large values can overflow strict parse into an empty result; verify against real data.")
if [ "$CARDINALITY" = plural ]; then
  MECH_BULLETS+=("Plural/list-all: no server-side filter (id is computed) — narrow client-side; test with set-membership, not a fixed index or count.")
  MECH_BULLETS+=("Read-after-write lag: create-then-list in one apply can transiently return 0 rows — require cassette replay before claiming it works.")
fi

RISK_MATERIAL=false
RISK_SUMMARY=""
LLM_BULLETS_JSON="[]"

# Static instructions live in prompts/risk-scan.md; the dynamic context (flags,
# report diagnostics, slice) is appended below, then the whole file is sent.
RISK_PROMPT_FILE="$(mktemp -t tfgen-risk.XXXXXX.txt)"
{
  cat "$PROMPTS_DIR/risk-scan.md"
  printf '\n\n---\nCONTEXT FOR THIS DATA SOURCE\n'
  printf 'Artifact: %s   Cardinality: %s\n' "$ARTIFACT_NAME" "$CARDINALITY"
  printf 'Read op: %s   Search op: %s\n' "${READ_OP:-none}" "${SEARCH_OP:-none}"
  printf 'Deterministic flags already recorded (do NOT repeat): '
  [ "${#MECH_BULLETS[@]}" -gt 0 ] && printf '%s | ' "${MECH_BULLETS[@]}"
  printf '\n\nGenerator report (summary + diagnostics):\n'
  jq -c '{summary, diagnostics: [.artifacts[]?.diagnostics[]?]}' "$REPORT" 2>/dev/null || true
  printf '\n\nSliced OpenAPI — the source of truth (may be truncated):\n'
  head -c 50000 "$SLICE"
  printf '\n\nGenerated Go — review this against the spec above (may be truncated):\n'
  head -c 40000 "datadog/fwprovider/data_source_datadog_${ARTIFACT_NAME}.go" 2>/dev/null || true
} >"$RISK_PROMPT_FILE"

if RISK_JSON="$(call_claude "$RISK_PROMPT_FILE")"; then
  # Coerce each field to its expected type. call_claude only guarantees the reply
  # is valid JSON, not its shape — an unexpected type (or a non-object reply) would
  # otherwise crash a later `jq --argjson` and, under `set -e`, abort with no result.
  RISK_MATERIAL="$(printf '%s' "$RISK_JSON" | jq -r 'try (if .material_risk==true then "true" else "false" end) catch "false"')"
  RISK_SUMMARY="$(printf '%s' "$RISK_JSON" | jq -r 'try (if (.risk_summary|type)=="string" then .risk_summary else "" end) catch ""')"
  LLM_BULLETS_JSON="$(printf '%s' "$RISK_JSON" | jq -c 'try (if (.reviewer_notes|type)=="array" then [.reviewer_notes[]|tostring] else [] end) catch []')"
  jlog "risk scan ok (material=$RISK_MATERIAL, $(printf '%s' "$LLM_BULLETS_JSON" | jq 'length') notes)"
else
  jlog "risk scan unavailable or unparseable — flagging manual review"
  MECH_BULLETS+=("Automated risk scan did not run — a reviewer must scan runtime risks manually.")
fi

# Merge deterministic flags + model notes into one list.
MECH_JSON="$(printf '%s\n' "${MECH_BULLETS[@]:-}" | jq -R . | jq -s '[.[] | select(length>0)]')"
ALL_BULLETS_JSON="$(jq -c -n --argjson a "$MECH_JSON" --argjson b "$LLM_BULLETS_JSON" '$a + $b')"

# ---------------------------------------------------------------------------
# Stage: pr-prose — model writes the How-to-test steps from prompts/pr-prose.md
# ---------------------------------------------------------------------------
STAGE="pr-prose"
TEST_FILE="datadog/tests/data_source_datadog_${ARTIFACT_NAME}_test.go"
TEST_FUNC="$(grep -oE 'func TestAcc[A-Za-z0-9_]+' "$TEST_FILE" 2>/dev/null | head -1 | sed 's/^func //' || true)"
[ -n "$TEST_FUNC" ] || TEST_FUNC="TestAccDatadog${ARTIFACT_NAME}DataSource"

HOWTO_MD=""
PROSE_PROMPT_FILE="$(mktemp -t tfgen-prose.XXXXXX.txt)"
{
  cat "$PROMPTS_DIR/pr-prose.md"
  printf '\n\n---\nCONTEXT\n'
  printf 'Artifact: %s   Cardinality: %s\n' "$ARTIFACT_NAME" "$CARDINALITY"
  printf 'Acceptance test function: %s\n' "$TEST_FUNC"
  printf 'Test file: %s\n' "$TEST_FILE"
} >"$PROSE_PROMPT_FILE"

if PROSE_JSON="$(call_claude "$PROSE_PROMPT_FILE")"; then
  # Type-coerced like the risk fields above, for the same set -e safety reason.
  HOWTO_MD="$(printf '%s' "$PROSE_JSON" | jq -r 'try (if (.how_to_test|type)=="string" then .how_to_test else "" end) catch ""')"
  [ -n "$HOWTO_MD" ] && jlog "pr prose ok"
fi
if [ -z "$HOWTO_MD" ]; then
  jlog "pr prose unavailable — using static testing steps"
  HOWTO_MD="Record once against the Frog org, then replay (replay is what CI runs):

\`\`\`bash
# Frog test-org creds
eval \"\$(dd-auth --domain frog.datadoghq.com --force-app-key --no-cache --output)\"
export DD_TEST_CLIENT_API_KEY=\"\$DD_API_KEY\" DD_TEST_CLIENT_APP_KEY=\"\$DD_APP_KEY\"

# record (writes the cassette + .freeze), then commit both from datadog/tests/cassettes/
make testacc RECORD=true TESTARGS='-run ${TEST_FUNC}'

# replay offline — only a green replay proves it works
make testacc RECORD=false TESTARGS='-run ${TEST_FUNC}'
\`\`\`"
fi

# ---------------------------------------------------------------------------
# Stage: PR body — templated deterministically, risk narrative injected
# ---------------------------------------------------------------------------
STAGE="pr-body"
TITLE="[$SERVICE] Add datadog_${ARTIFACT_NAME} data source"
OPS_STR="$READ_OP"; [ -n "$SEARCH_OP" ] && OPS_STR="${OPS_STR:+$OPS_STR, }$SEARCH_OP"

RISK_CALLOUT=""
if [ "$RISK_MATERIAL" = true ] && [ -n "$RISK_SUMMARY" ]; then
  RISK_CALLOUT="> ⚠️ **Merge risks flagged — read before approving.** ${RISK_SUMMARY}
> Details under \"Reviewer notes / risks\" below.

"
fi

# Overwrite intro sentence carries its own leading newline so the additive case
# leaves no blank line in the paragraph.
OVERWRITE_LINE=""
if [ -n "$OVERWRITES" ]; then
  OVERWRITE_LINE="
The hand-written \`${OVERWRITES}\` it replaces was removed from \`framework_provider.go\`'s \`Datasources\` slice."
fi

# Build the Generated-files list so empty (additive-case) pieces add no blank bullet.
GEN_LIST="- \`datadog/fwprovider/data_source_datadog_${ARTIFACT_NAME}.go\`
- \`datadog/tests/data_source_datadog_${ARTIFACT_NAME}_test.go\`
- \`datadog/fwprovider/datasources_generated.go\` — registers the new constructor"
if [ -n "$OVERWRITES" ]; then
  GEN_LIST="${GEN_LIST}
- \`datadog/fwprovider/framework_provider.go\` (updated) — retired constructor removed"
fi
GEN_LIST="${GEN_LIST}
- \`docs/data-sources/${ARTIFACT_NAME}.md\` (created via \`make docs\`)"

RISK_BULLETS_MD="$(printf '%s' "$ALL_BULLETS_JSON" | jq -r '.[] | "- " + .')"
[ -z "$RISK_BULLETS_MD" ] && RISK_BULLETS_MD="- No material runtime risks flagged by the automated scan."
WARN_MD="$(printf '%s' "$WARN_DIAGS" | jq -r '.[]? | "- generator " + .severity + ": " + .message')"

PR_BODY_FILE="$(mktemp -t tfgen-pr-body.XXXXXX.md)"
cat >"$PR_BODY_FILE" <<EOF
> ℹ️ **This PR is part of a new project that auto-generates Terraform provider data
> sources to increase coverage.** The data source code is generated **deterministically
> by tfgen from an annotated OpenAPI spec, without the use of LLMs**. This PR is
> **reviewed via AI, though human review is still necessary**. If you decide to use
> this generated data source, please **review it thoroughly and test it** before relying
> on it — see the verification note and testing guide below.

> ⚠️ **This PR contains auto-generated code and must be verified before merging.** Do not
> merge until the acceptance test has been recorded and replays green against the Frog org
> (see "How to test"). A clean build and a green generator report do **not** prove runtime
> correctness.

${RISK_CALLOUT}## ${ARTIFACT_NAME} data source (generator-v2)

Generated by tfgen from an annotated slice of the Datadog v2 OpenAPI spec
(operations: \`${OPS_STR}\`; cardinality: \`${CARDINALITY}\`).
Registered in \`datasources_generated.go\`'s \`generatedDatasources\` slice, which
\`framework_provider.go\` appends alongside the hand-written \`Datasources\` — tfgen owns that
file, so it is not hand-edited.${OVERWRITE_LINE}

**Spec provenance:** ${SPEC_SOURCE}$([ -n "$SPEC_SHA" ] && echo " @ ${SPEC_SHA}")$([ -n "$SPEC_HASH" ] && echo " (slice hash \`${SPEC_HASH}\`)")

### Generated
${GEN_LIST}

### Cardinality
${CARDINALITY}

### Test / cassette
- Acceptance test: \`${TEST_FUNC}\` in \`${TEST_FILE}\`
- Cassette: scaffold, not yet recorded

### Reviewer notes / risks
${RISK_BULLETS_MD}
${WARN_MD}

### How to test
${HOWTO_MD}

---
> 🚧 **This Terraform data source generation is still under development.** Reach out to
> **#api-platform** with any questions.
EOF
jlog "PR body drafted: $PR_BODY_FILE"

# ---------------------------------------------------------------------------
# Stage: branch + commit
# ---------------------------------------------------------------------------
STAGE="commit"
git add "datadog/fwprovider/data_source_datadog_${ARTIFACT_NAME}.go" \
        "datadog/tests/data_source_datadog_${ARTIFACT_NAME}_test.go" \
        "datadog/fwprovider/datasources_generated.go" \
        "$DOCS_FILE" >&2
[ -n "$OVERWRITES" ] && git add "datadog/fwprovider/framework_provider.go" >&2
git commit -m "$TITLE (generated)" >&2 || die "git commit failed"
COMMITTED=1
jlog "committed on $BRANCH (from $BASE_REF)"

# ---------------------------------------------------------------------------
# Stage: PR (always a draft) — or stop here with --no-pr
# ---------------------------------------------------------------------------
PR_URL=""
if [ "$NO_PR" -eq 1 ]; then
  jlog "--no-pr set; stopping after commit"
else
  STAGE="pr"
  git push -u origin "$BRANCH" >&2 || die "git push failed"
  # CI wants a changelog/* label, but a repo/fork that lacks it shouldn't sink the
  # PR after we've pushed — attach it only if it exists, and warn otherwise.
  PR_LABEL="changelog/feature"
  pr_args=(--draft --base "$BASE" --head "$BRANCH" --title "$TITLE" --body-file "$PR_BODY_FILE")
  if gh label list --json name --jq '.[].name' 2>/dev/null | grep -Fxq "$PR_LABEL"; then
    pr_args+=(--label "$PR_LABEL")
  else
    jlog "WARNING: label '$PR_LABEL' not on the repo; opening PR without it (CI's changelog check may fail)"
  fi
  PR_URL="$(gh pr create "${pr_args[@]}" 2>&1)" || die "gh pr create failed: $PR_URL"
  jlog "draft PR: $PR_URL"
fi

# ---------------------------------------------------------------------------
# Result
# ---------------------------------------------------------------------------
STAGE="done"
METRICS="$(metrics_json)"
RESULT_JSON="$(jq -n \
  --arg status succeeded \
  --arg artifact_name "$ARTIFACT_NAME" \
  --arg cardinality "$CARDINALITY" \
  --arg service "$SERVICE" \
  --arg read_op "$READ_OP" --arg search_op "$SEARCH_OP" \
  --arg overwrites "$OVERWRITES" \
  --arg branch "$BRANCH" --arg base "$BASE" \
  --arg spec_source "$SPEC_SOURCE" --arg spec_sha "$SPEC_SHA" --arg spec_hash "$SPEC_HASH" \
  --arg report "$REPORT" --arg pr_url "$PR_URL" \
  --argjson material_risk "$RISK_MATERIAL" \
  --arg risk_summary "$RISK_SUMMARY" \
  --argjson risks "$ALL_BULLETS_JSON" \
  --argjson changed_files "$CHANGED_JSON" \
  --argjson warnings "$WARN_DIAGS" \
  --argjson metrics "$METRICS" \
  '{status:$status, verified:false, artifact_name:$artifact_name, cardinality:$cardinality,
    service:$service, operations:{read:$read_op, search:$search_op}, overwrites:$overwrites,
    branch:$branch, base:$base, spec:{source:$spec_source, sha:$spec_sha, slice_hash:$spec_hash},
    report_path:$report, pr_url:$pr_url, material_risk:$material_risk, risk_summary:$risk_summary,
    risks:$risks, changed_files:$changed_files, generator_warnings:$warnings, metrics:$metrics}')"

jlog "$(printf '%s' "$METRICS" | jq -r '"done in \(.runtime_seconds)s | claude: \(.claude_calls) calls, $\(.claude_cost_usd), \(.claude_input_tokens) in / \(.claude_output_tokens) out tokens"')"
emit_result "$RESULT_JSON"
