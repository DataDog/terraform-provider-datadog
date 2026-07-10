#!/usr/bin/env bash
# Headless spine for keeping the provider in lockstep with an already-annotated
# Datadog v2 OpenAPI spec: generate every annotated data source, retire the ones
# whose annotation is gone, and fan each impacted artifact out to its own draft PR.
#
# Two-part flow. First, ONE expensive run in a staging branch proves the whole set
# together: reconcile (generate-all + retire orphans) -> make docs -> make build.
# Then a cheap per-artifact fan-out cuts a branch from base, re-emits just that
# artifact (or retires it), and opens a draft PR. Build is verified in aggregate;
# each PR's own CI is the per-artifact safety net.
#
# Safety model: nothing runs on a dirty tree, every PR is a draft, retirement is
# gated by the generator's cassette check (adopted data sources are flagged, never
# deleted), and --dry-run / --max-prs guard against a runaway fan-out.
#
# stdout carries ONLY the final JSON result. All human logs go to stderr.
# Exit 0 = success; nonzero = a gate or setup failure (with JSON on stdout).

set -euo pipefail

# Associative arrays, mapfile and ${arr[@]} under set -u need bash >= 4.
if [[ "${BASH_VERSINFO:-0}" -lt 4 ]]; then
  echo "generate_batch.sh requires bash >= 4 (found ${BASH_VERSION:-?}); on macOS: brew install bash" >&2
  exit 2
fi

# Resolve this script's dir before we cd, and stash the prompt files OUTSIDE the
# tree: the staging/fan-out branches reset the working tree, which would drop
# prompts/ if the base branch doesn't carry it.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROMPTS_DIR="$(mktemp -d -t tfgen-batch-prompts.XXXXXX)"
cp "$SCRIPT_DIR/prompts/"*.md "$PROMPTS_DIR/" 2>/dev/null || true
CAPTURE_DIR="$(mktemp -d -t tfgen-batch-docs.XXXXXX)"

cleanup() {
  [[ -n "${PROMPTS_DIR:-}" ]] && rm -rf "$PROMPTS_DIR" 2>/dev/null || true
  [[ -n "${CAPTURE_DIR:-}" ]] && rm -rf "$CAPTURE_DIR" 2>/dev/null || true
  rm -f "${CLAUDE_COST_FILE:-}" "${REPORT:-}" "${NAME_SERVICE_FILE:-}" 2>/dev/null || true
  # Best-effort return to the branch we started on; never touch pushed branches.
  [[ -n "${ORIG_BRANCH:-}" ]] && git checkout "$ORIG_BRANCH" >/dev/null 2>&1 || true
  [[ -n "${STAGING_BRANCH:-}" ]] && git branch -D "$STAGING_BRANCH" >/dev/null 2>&1 || true
  return 0
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Logging + structured failure
# ---------------------------------------------------------------------------
STAGE="init"
jlog() { printf '%s\n' "[$STAGE] $*" >&2; }

emit_result() {
  printf '%s\n' "$1"
  [[ -n "${OUTPUT_JSON:-}" ]] && printf '%s\n' "$1" >"$OUTPUT_JSON"
  return 0
}

# die <message> — restore the working tree to base, then emit a failure result.
die() {
  local msg="$1"
  jlog "FATAL: $msg"
  restore_base
  emit_result "$(jq -n --arg status failed --arg stage "$STAGE" --arg error "$msg" \
    --arg base "${BASE:-}" --argjson metrics "$(metrics_json)" \
    '{status:$status, stage:$stage, error:$error, base:$base, metrics:$metrics}')"
  exit 1
}

# restore_base — discard everything this run generated and get back to a clean
# base checkout, so a failure never strands the tree mid-generation. The
# destructive reset is gated on MUTATED so a precondition failure (e.g. a dirty
# tree the user meant to keep) never resets work this script did not create.
restore_base() {
  if [[ "${MUTATED:-0}" -eq 1 ]]; then
    git reset --hard >/dev/null 2>&1 || true
    git clean -fdq datadog/fwprovider datadog/tests docs/data-sources >/dev/null 2>&1 || true
  fi
  if [[ -n "${ORIG_BRANCH:-}" ]]; then
    git checkout -f "$ORIG_BRANCH" >/dev/null 2>&1 || true
  fi
  if [[ -n "${STAGING_BRANCH:-}" ]]; then
    git branch -D "$STAGING_BRANCH" >/dev/null 2>&1 || true
    STAGING_BRANCH=""
  fi
}

usage() {
  cat >&2 <<'EOF'
Usage: generate_batch.sh --spec PATH [flags]

Keeps the provider's generated data sources in lockstep with an already-annotated
v2 OpenAPI spec: one PR per created / updated / retired artifact.

Required:
  --spec PATH        full v2 OAS, already annotated with x-datadog-tf-generator

Optional:
  --base BRANCH      branch each PR targets and is built from (default: current branch)
  --max-prs N        abort before fan-out if more than N PRs would open (default: 25)
  --dry-run          verify + print the plan; open no branches or PRs
  --no-ai            skip all LLM calls; PR bodies use deterministic notes only
  --no-pr            commit each artifact on its branch but do not push or open a PR
  --output-json PATH also write the final batch JSON here
EOF
  exit 2
}

CLAUDE_COST_FILE=""

# call_claude <prompt-file> — print the model's JSON reply (fences stripped) or
# return 1 if claude is missing/errors/returns non-JSON. Honors --no-ai.
call_claude() {
  [[ "${NO_AI:-0}" = 1 ]] && return 1
  local pf="$1" raw result scratch to
  command -v claude >/dev/null 2>&1 || return 1
  to=""; for t in timeout gtimeout; do command -v "$t" >/dev/null 2>&1 && { to="$t 180"; break; }; done
  scratch="$(mktemp -d -t tfgen-batch-claude.XXXXXX)"
  raw="$( cd "$scratch" && $to claude -p "$(cat "$pf")" --strict-mcp-config --output-format json --max-turns 1 2>/dev/null || true )"
  rm -rf "$scratch" 2>/dev/null || true
  [[ -n "$CLAUDE_COST_FILE" ]] && printf '%s\n' "$raw" | jq -c \
    '{cost: (.total_cost_usd // .cost_usd // 0), in: ((.usage.input_tokens // 0) + (.usage.cache_read_input_tokens // 0) + (.usage.cache_creation_input_tokens // 0)), out: (.usage.output_tokens // 0)}' \
    >>"$CLAUDE_COST_FILE" 2>/dev/null || true
  result="$(printf '%s' "$raw" | jq -r 'if type=="object" then (.result // "") else . end' 2>/dev/null || true)"
  result="$(printf '%s' "$result" | sed -e 's/^```json//' -e 's/^```//' -e 's/```$//')"
  printf '%s' "$result" | jq -e . >/dev/null 2>&1 || return 1
  printf '%s' "$result"
}

metrics_json() {
  local cost=0 intok=0 outtok=0 calls=0
  if [[ -n "$CLAUDE_COST_FILE" && -s "$CLAUDE_COST_FILE" ]]; then
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
SPEC="" BASE="" MAX_PRS=25 DRY_RUN=0 NO_AI=0 NO_PR=0 OUTPUT_JSON=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --spec)        SPEC="${2:-}"; shift 2 ;;
    --base)        BASE="${2:-}"; shift 2 ;;
    --max-prs)     MAX_PRS="${2:-}"; shift 2 ;;
    --dry-run)     DRY_RUN=1; shift ;;
    --no-ai)       NO_AI=1; shift ;;
    --no-pr)       NO_PR=1; shift ;;
    --output-json) OUTPUT_JSON="${2:-}"; shift 2 ;;
    -h|--help)     usage ;;
    *) echo "unknown flag: $1" >&2; usage ;;
  esac
done

# ---------------------------------------------------------------------------
# Stage: preconditions
# ---------------------------------------------------------------------------
STAGE="preconditions"

[[ -n "$SPEC" ]] || die "missing --spec (a full, already-annotated v2 OAS)"
[[ -f "$SPEC" ]] || die "--spec not found: $SPEC"
[[ "$MAX_PRS" =~ ^[0-9]+$ ]] || die "--max-prs must be a non-negative integer, got '$MAX_PRS'"

for tool in git gh python3 make jq; do
  command -v "$tool" >/dev/null 2>&1 || die "required tool not found: $tool"
done
python3 -c 'import yaml' 2>/dev/null || die "python3 PyYAML not installed (pip install pyyaml)"
if [[ "$DRY_RUN" -eq 0 && "$NO_PR" -eq 0 ]]; then
  gh auth status >/dev/null 2>&1 || die "gh is not authenticated (run: gh auth login)"
fi

# Repo root = the dir holding the generator; search upward so this works from a subdir.
find_root() {
  local d; d="$(pwd)"
  while [[ "$d" != "/" ]]; do
    [[ -f "$d/.generator-v2/cmd/tfgen/main.go" ]] && { printf '%s\n' "$d"; return 0; }
    d="$(dirname "$d")"
  done
  return 1
}
ROOT="$(find_root)" || die "not inside a terraform-provider-datadog checkout (no .generator-v2/cmd/tfgen)"
# --spec may be given relative to the caller's cwd; resolve it before we cd.
SPEC="$(cd "$(dirname "$SPEC")" && pwd)/$(basename "$SPEC")"
cd "$ROOT"
jlog "repo root: $ROOT"

[[ -z "$(git status --porcelain)" ]] || die "working tree is dirty; commit or stash first"

ORIG_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [[ -z "$BASE" ]]; then
  [[ "$ORIG_BRANCH" != "HEAD" ]] || die "detached HEAD — pass --base explicitly"
  BASE="$ORIG_BRANCH"
fi

# Resolve the ref we branch from and target. Prefer origin's copy (the PR target)
# when reachable, so per-artifact diffs are clean; fall back to the local branch.
BASE_REF=""
if git fetch --quiet origin "$BASE" 2>/dev/null; then
  BASE_REF="origin/$BASE"
elif git rev-parse --verify --quiet "refs/heads/$BASE" >/dev/null; then
  BASE_REF="$BASE"
else
  die "base branch '$BASE' not found on origin or locally"
fi
# GitHub can only target a branch it has.
if [[ "$DRY_RUN" -eq 0 && "$NO_PR" -eq 0 ]]; then
  git ls-remote --exit-code --heads origin "$BASE" >/dev/null 2>&1 \
    || die "base '$BASE' does not exist on origin; push it first or use --dry-run/--no-pr"
fi
jlog "base: $BASE ($BASE_REF)"

# ---------------------------------------------------------------------------
# Stage: tfgen-build
# ---------------------------------------------------------------------------
STAGE="tfgen-build"
make tfgen-build >&2 || die "make tfgen-build failed"
[[ -x bin/tfgen ]] || die "bin/tfgen missing after build"

# Name -> service tag, for the CI-enforced [prefix] in each PR title. Built from
# the annotated ops; retired artifacts (no longer annotated) fall back to "core".
NAME_SERVICE_FILE="$(mktemp -t tfgen-batch-svc.XXXXXX)"
python3 - "$SPEC" >"$NAME_SERVICE_FILE" <<'PY' || true
import sys, yaml
spec = yaml.safe_load(open(sys.argv[1]))
for item in (spec.get("paths") or {}).values():
    if not isinstance(item, dict):
        continue
    for node in item.values():
        if isinstance(node, dict) and "x-datadog-tf-generator" in node:
            ext = node.get("x-datadog-tf-generator") or {}
            name = ext.get("artifact_name")
            tags = node.get("tags") or []
            if name and tags:
                print(f"{name}\t{tags[0].lower().replace(' ', '-')}")
PY

service_for() {
  local name="$1" svc
  svc="$(awk -F'\t' -v n="$name" '$1==n {print $2; exit}' "$NAME_SERVICE_FILE" 2>/dev/null || true)"
  printf '%s' "${svc:-core}"
}

# ---------------------------------------------------------------------------
# Stage: verify — the single expensive reconcile + docs + build, on a staging branch
# ---------------------------------------------------------------------------
STAGE="verify"
STAGING_BRANCH="batch-staging-$$"
git checkout -b "$STAGING_BRANCH" "$BASE_REF" >&2 || die "failed to create staging branch from $BASE_REF"
# Past this point the tree carries our generated changes, so restore_base may reset it.
MUTATED=1

REPORT="$(mktemp -t tfgen-batch-report.XXXXXX.json)"
if ! ./bin/tfgen generate --spec "$SPEC" --emit-tests --reconcile --report "$REPORT" >&2; then
  jlog "tfgen exited nonzero (gate below inspects the report)"
fi
[[ -s "$REPORT" ]] || die "tfgen wrote no report"

# ---------------------------------------------------------------------------
# Stage: gate — stop before fan-out on any failed artifact / error diagnostic
# ---------------------------------------------------------------------------
STAGE="gate"
FAILED="$(jq -r '.summary.failed // 0' "$REPORT")"
ERROR_DIAGS="$(jq -c '[.artifacts[]?.diagnostics[]? | select(.severity=="error")]' "$REPORT")"
if [[ "$FAILED" != "0" || "$ERROR_DIAGS" != "[]" ]]; then
  jlog "GATE FAILED — failed=$FAILED errors=$ERROR_DIAGS"
  restore_base
  emit_result "$(jq -n --arg status failed --arg stage gate \
    --argjson failed "$FAILED" --argjson errors "$ERROR_DIAGS" \
    --argjson metrics "$(metrics_json)" \
    '{status:$status, stage:$stage, error:"generation gate failed", failed:$failed, error_diagnostics:$errors, metrics:$metrics}')"
  exit 1
fi
SPEC_HASH="$(jq -r '.spec_hash // empty' "$REPORT")"
jlog "gate passed (failed=0, no error diagnostics)"

# Impacted artifact names, by action. Data-source entries only (drop the _test.go
# scaffold entries and any registry-only orphan with no file path).
ds_names() {
  jq -r --arg st "$1" \
    '.artifacts[] | select((.path//"")|endswith(".go")) | select((.path//"")|endswith("_test.go")|not) | select(.status==$st) | .name' \
    "$REPORT" | sort -u
}
mapfile -t CREATED  < <(ds_names created)
mapfile -t UPDATED  < <(ds_names updated)
mapfile -t RETIRED  < <(ds_names retired)
mapfile -t BLOCKED  < <(ds_names retire_blocked)

# ---------------------------------------------------------------------------
# Stage: docs — regenerate, keep only impacted pages, revert unrelated drift
# ---------------------------------------------------------------------------
STAGE="docs"
make docs >&2 || die "make docs failed"

# Pages we intend to change: created/updated stay (present), retired stay deleted.
declare -A KEEP_DOC=()
for n in "${CREATED[@]}" "${UPDATED[@]}" "${RETIRED[@]}"; do
  [[ -n "$n" ]] && KEEP_DOC["docs/data-sources/${n}.md"]=1
done
while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  status="${line:0:2}"; path="${line:3}"
  path="${path#\"}"; path="${path%\"}"
  [[ -n "${KEEP_DOC[$path]:-}" ]] && continue
  if [[ "$status" == "??" ]]; then
    rm -f "$path" 2>/dev/null || true
  else
    git checkout -- "$path" >/dev/null 2>&1 || true
  fi
  jlog "reverted unrelated doc drift: $path"
done < <(git status --porcelain -- docs/)

# Every created/updated artifact must have produced a docs page.
for n in "${CREATED[@]}" "${UPDATED[@]}"; do
  [[ -z "$n" ]] && continue
  [[ -f "docs/data-sources/${n}.md" ]] || die "make docs produced no docs/data-sources/${n}.md — registration for '$n' did not take"
done

# ---------------------------------------------------------------------------
# Stage: build — prove the whole impacted set compiles together (once)
# ---------------------------------------------------------------------------
STAGE="build"
make build >&2 || die "make build failed (the generated set does not compile together)"
jlog "docs + build clean for the full set"

# Capture the created/updated docs before we discard the staging tree; the
# per-artifact fan-out copies them back rather than re-running make docs N times.
for n in "${CREATED[@]}" "${UPDATED[@]}"; do
  [[ -z "$n" ]] && continue
  cp "docs/data-sources/${n}.md" "$CAPTURE_DIR/${n}.md" 2>/dev/null || true
done

# Restore base: the staging tree has done its job (verify + capture). restore_base
# discards the generated mutations, returns to ORIG_BRANCH and drops the staging branch.
restore_base

# ---------------------------------------------------------------------------
# Stage: plan — count PRs, enforce the cap, short-circuit on --dry-run
# ---------------------------------------------------------------------------
STAGE="plan"
N_CREATED="${#CREATED[@]}"; N_UPDATED="${#UPDATED[@]}"
N_RETIRED="${#RETIRED[@]}"; N_BLOCKED="${#BLOCKED[@]}"
PR_COUNT=$((N_CREATED + N_UPDATED + N_RETIRED))
jlog "plan: create=$N_CREATED update=$N_UPDATED retire=$N_RETIRED blocked=$N_BLOCKED (PRs=$PR_COUNT)"

if [[ "$PR_COUNT" -gt "$MAX_PRS" ]]; then
  die "batch would open $PR_COUNT PRs, over --max-prs $MAX_PRS; raise --max-prs or narrow the spec"
fi

blocked_json() { printf '%s\n' "${BLOCKED[@]}" | jq -R . | jq -s '[.[] | select(length>0)]'; }
counts_json() {
  jq -n --argjson c "$N_CREATED" --argjson u "$N_UPDATED" --argjson r "$N_RETIRED" \
        --argjson b "$N_BLOCKED" --argjson unchanged "$(jq -r '.summary.unchanged // 0' "$REPORT")" \
    '{created:$c, updated:$u, retired:$r, retire_blocked:$b, unchanged:$unchanged}'
}

if [[ "$DRY_RUN" -eq 1 ]]; then
  STAGE="done"
  emit_result "$(jq -n \
    --arg status planned --arg base "$BASE" --arg spec "$SPEC" --arg spec_hash "$SPEC_HASH" \
    --argjson counts "$(counts_json)" \
    --argjson created "$(printf '%s\n' "${CREATED[@]}" | jq -R . | jq -s '[.[]|select(length>0)]')" \
    --argjson updated "$(printf '%s\n' "${UPDATED[@]}" | jq -R . | jq -s '[.[]|select(length>0)]')" \
    --argjson retired "$(printf '%s\n' "${RETIRED[@]}" | jq -R . | jq -s '[.[]|select(length>0)]')" \
    --argjson blocked "$(blocked_json)" \
    --argjson metrics "$(metrics_json)" \
    '{status:$status, dry_run:true, base:$base, spec:{path:$spec, hash:$spec_hash}, counts:$counts,
      plan:{created:$created, updated:$updated, retired:$retired, retire_blocked:$blocked}, metrics:$metrics}')"
  jlog "dry-run: no branches or PRs opened"
  exit 0
fi

# ---------------------------------------------------------------------------
# Fan-out — one branch + draft PR per artifact. Fail-slow: a bad artifact is
# recorded and skipped, never aborting the batch.
# ---------------------------------------------------------------------------

DISCLAIMER_TOP='> ℹ️ **This PR is part of a project that auto-generates Terraform provider data
> sources to increase coverage.** The code is generated **deterministically by tfgen from an
> annotated OpenAPI spec, without the use of LLMs**, and **reviewed via AI, though human review
> is still necessary**. If you use this data source, **review it thoroughly and test it** first.'
DISCLAIMER_VERIFY='> ⚠️ **This PR contains auto-generated code and must be verified before merging.** Do not
> merge until the acceptance test is recorded and replays green against the Frog org. A clean
> build and a green generator report do **not** prove runtime correctness.'
DISCLAIMER_FOOT='---
> 🚧 **This Terraform data source generation is still under development.** Reach out to
> **#api-platform** with any questions.'

STATIC_HOWTO() {
  local fn="$1"
  cat <<EOF
Record once against the Frog org, then replay (replay is what CI runs):

\`\`\`bash
eval "\$(dd-auth --domain frog.datadoghq.com --force-app-key --no-cache --output)"
export DD_TEST_CLIENT_API_KEY="\$DD_API_KEY" DD_TEST_CLIENT_APP_KEY="\$DD_APP_KEY"

make testacc RECORD=true  TESTARGS='-run ${fn}'   # records the cassette + .freeze
make testacc RECORD=false TESTARGS='-run ${fn}'   # replay offline — only green proves it works
\`\`\`
EOF
}

# process_artifact <action> <name> — action is created|updated|retired. Prints one
# per-artifact result JSON to stdout; all logs go to stderr. Returns nonzero on a
# handled failure (already recorded in the JSON) so the caller can tally it.
process_artifact() {
  set +e  # fail-slow: we check each step explicitly rather than aborting
  local action="$1" name="$2"
  local svc branch pr_url="" err="" report diff_ctx
  local committed=0 branch_created=0
  svc="$(service_for "$name")"

  local title verb
  case "$action" in
    created) verb="Add";    branch="generate/datadog_${name}_datasource" ;;
    updated) verb="Update"; branch="generate/datadog_${name}_datasource" ;;
    retired) verb="Retire"; branch="retire/datadog_${name}_datasource" ;;
    *) verb="?"; branch="generate/datadog_${name}_datasource" ;;
  esac
  title="[$svc] $verb datadog_${name} data source"

  pa_result() {  # <status> [extra-json]
    local extra="${2:-}"; [[ -z "$extra" ]] && extra='{}'
    jq -n --arg status "$1" --arg action "$action" --arg name "$name" \
          --arg branch "$branch" --arg pr_url "$pr_url" --arg error "$err" \
          --argjson extra "$extra" \
      '{artifact_name:$name, action:$action, status:$status, branch:$branch,
        pr_url:(if $pr_url=="" then null else $pr_url end),
        error:(if $error=="" then null else $error end)} + $extra'
  }
  pa_fail() {  # <message> — undo only what we created here, then report
    err="$1"; jlog "artifact '$name' failed: $err"
    if [[ "$branch_created" -eq 1 ]]; then
      git reset --hard >/dev/null 2>&1 || true
      git clean -fdq datadog/fwprovider datadog/tests docs/data-sources >/dev/null 2>&1 || true
      git checkout -f "$ORIG_BRANCH" >/dev/null 2>&1 || true
      [[ "$committed" -eq 0 ]] && git branch -D "$branch" >/dev/null 2>&1 || true
    fi
    pa_result failed
    return 1
  }

  # Names come from tfgen's schema/gate-validated report, but re-check here before
  # any name becomes a git ref or path, so a drifted report can never inject a
  # traversal or a bad ref (fail closed; branch_created is still 0).
  if [[ ! "$name" =~ ^[a-z][a-z0-9_]*$ ]]; then
    pa_fail "invalid artifact name '$name'"; return 1
  fi

  # Branch collision — skip rather than clobber (branch_created stays 0, so pa_fail
  # leaves the pre-existing branch untouched).
  if git rev-parse --verify --quiet "refs/heads/$branch" >/dev/null \
     || git ls-remote --exit-code --heads origin "$branch" >/dev/null 2>&1; then
    pa_fail "branch '$branch' already exists (local or origin)"; return 1
  fi

  git checkout -b "$branch" "$BASE_REF" >&2 || { pa_fail "could not create branch from $BASE_REF"; return 1; }
  branch_created=1

  report="$(mktemp -t tfgen-batch-art.XXXXXX.json)"
  if [[ "$action" == "retired" ]]; then
    ./bin/tfgen generate --retire "$name" --report "$report" >&2 \
      || { rm -f "$report"; pa_fail "tfgen --retire failed"; return 1; }
  else
    ./bin/tfgen generate --spec "$SPEC" --include "$name" --emit-tests --report "$report" >&2 \
      || { rm -f "$report"; pa_fail "tfgen --include failed"; return 1; }
    # Any error diagnostic here means the scoped re-emit disagrees with the batch run.
    if [[ "$(jq -r '.summary.failed // 0' "$report")" != "0" ]]; then
      rm -f "$report"; pa_fail "scoped re-emit reported a failed artifact"; return 1
    fi
    cp "$CAPTURE_DIR/${name}.md" "docs/data-sources/${name}.md" 2>/dev/null \
      || { rm -f "$report"; pa_fail "captured docs page missing for '$name'"; return 1; }
  fi

  # Whitelist: assert only this artifact's files changed.
  local go_f="datadog/fwprovider/data_source_datadog_${name}.go"
  local test_f="datadog/tests/data_source_datadog_${name}_test.go"
  local doc_f="docs/data-sources/${name}.md"
  local reg_f="datadog/fwprovider/datasources_generated.go"
  declare -A allow=()
  allow["$go_f"]=1; allow["$test_f"]=1; allow["$doc_f"]=1; allow["$reg_f"]=1
  # A created/updated artifact with overwrites also edits framework_provider.go.
  [[ "$action" != "retired" ]] && allow["datadog/fwprovider/framework_provider.go"]=1
  local unexpected=()
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    local p="${line:3}"; p="${p#\"}"; p="${p%\"}"
    [[ -n "${allow[$p]:-}" ]] || unexpected+=("$p")
  done < <(git status --porcelain)
  if [[ "${#unexpected[@]}" -gt 0 ]]; then
    rm -f "$report"; pa_fail "files changed outside the whitelist: ${unexpected[*]}"; return 1
  fi

  # PR body.
  local test_func risk_bullets_md="" howto_md material_risk=false risk_summary="" risk_callout=""
  local warn_md changed_json body_file
  changed_json="$(git status --porcelain | sed 's/^...//' | jq -R . | jq -s .)"
  warn_md="$(jq -r '[.artifacts[]?.diagnostics[]? | select(.severity!="error")] | .[]? | "- generator " + .severity + ": " + .message' "$report" 2>/dev/null || true)"

  if [[ "$action" == "retired" ]]; then
    body_file="$(build_retire_body "$name")"
  else
    test_func="$(grep -oE 'func TestAcc[A-Za-z0-9_]+' "$test_f" 2>/dev/null | head -1 | sed 's/^func //' || true)"
    [[ -n "$test_func" ]] || test_func="TestAccDatadog${name}DataSource"

    # Risk scan over the generated Go (+ its diff for an update); no spec slice in batch.
    diff_ctx=""
    [[ "$action" == "updated" ]] && diff_ctx="$(git diff "$BASE_REF" -- "$go_f" 2>/dev/null | head -c 20000 || true)"
    local risk_json="" llm_bullets="[]"
    local risk_prompt; risk_prompt="$(mktemp -t tfgen-batch-risk.XXXXXX.txt)"
    {
      cat "$PROMPTS_DIR/risk-scan.md"
      printf '\n\n---\nCONTEXT FOR THIS DATA SOURCE (batch mode: no spec slice — review the generated Go)\n'
      printf 'Artifact: %s   Action: %s\n' "$name" "$action"
      printf '\nGenerator report diagnostics:\n'
      jq -c '[.artifacts[]?.diagnostics[]?]' "$report" 2>/dev/null || true
      [[ -n "$diff_ctx" ]] && { printf '\n\nDiff of the generated Go vs base:\n'; printf '%s' "$diff_ctx"; }
      printf '\n\nGenerated Go (review against the diagnostics; may be truncated):\n'
      head -c 40000 "$go_f" 2>/dev/null || true
    } >"$risk_prompt"
    if risk_json="$(call_claude "$risk_prompt")"; then
      material_risk="$(printf '%s' "$risk_json" | jq -r 'try (if .material_risk==true then "true" else "false" end) catch "false"')"
      risk_summary="$(printf '%s' "$risk_json" | jq -r 'try (if (.risk_summary|type)=="string" then .risk_summary else "" end) catch ""')"
      llm_bullets="$(printf '%s' "$risk_json" | jq -c 'try (if (.reviewer_notes|type)=="array" then [.reviewer_notes[]|tostring] else [] end) catch []')"
    elif [[ "$NO_AI" -eq 1 ]]; then
      jlog "risk scan skipped (--no-ai)"
    else
      jlog "risk scan unavailable for '$name' — flagging manual review"
    fi
    rm -f "$risk_prompt"

    # Deterministic bullets = generator diagnostics; merge with any model notes.
    local mech_bullets; mech_bullets="$(jq -c '[.artifacts[]?.diagnostics[]? | select(.severity!="error") | .message]' "$report" 2>/dev/null || echo '[]')"
    [[ "$NO_AI" -eq 1 ]] && mech_bullets="$(printf '%s' "$mech_bullets" | jq -c '. + ["Automated risk scan skipped (--no-ai) — a reviewer must scan runtime risks manually."]')"
    local all_bullets; all_bullets="$(jq -c -n --argjson a "$mech_bullets" --argjson b "$llm_bullets" '$a + $b')"
    risk_bullets_md="$(printf '%s' "$all_bullets" | jq -r '.[] | "- " + .')"
    [[ -z "$risk_bullets_md" ]] && risk_bullets_md="- No material runtime risks flagged."

    # How-to-test (LLM, else static).
    howto_md=""
    if [[ "$NO_AI" -eq 0 ]]; then
      local prose_prompt; prose_prompt="$(mktemp -t tfgen-batch-prose.XXXXXX.txt)"
      { cat "$PROMPTS_DIR/pr-prose.md"; printf '\n\n---\nCONTEXT\nArtifact: %s\nAcceptance test function: %s\nTest file: %s\n' "$name" "$test_func" "$test_f"; } >"$prose_prompt"
      local prose_json
      if prose_json="$(call_claude "$prose_prompt")"; then
        howto_md="$(printf '%s' "$prose_json" | jq -r 'try (if (.how_to_test|type)=="string" then .how_to_test else "" end) catch ""')"
      fi
      rm -f "$prose_prompt"
    fi
    [[ -z "$howto_md" ]] && howto_md="$(STATIC_HOWTO "$test_func")"

    [[ "$material_risk" == true && -n "$risk_summary" ]] && risk_callout="> ⚠️ **Merge risks flagged — read before approving.** ${risk_summary}

"
    body_file="$(build_generate_body "$name" "$svc" "$action" "$test_func" "$test_f" "$risk_callout" "$risk_bullets_md" "$warn_md" "$howto_md")"
  fi
  rm -f "$report"

  git add -A "$go_f" "$test_f" "$reg_f" "$doc_f" datadog/fwprovider/framework_provider.go >/dev/null 2>&1 || true
  git commit -m "$title (generated)" >&2 || { rm -f "$body_file"; pa_fail "git commit failed"; return 1; }
  committed=1

  if [[ "$NO_PR" -eq 1 ]]; then
    jlog "--no-pr: committed '$name' on $branch, not pushing"
  else
    git push -u origin "$branch" >&2 || { rm -f "$body_file"; pa_fail "git push failed"; return 1; }
    local pr_args=(--draft --base "$BASE" --head "$branch" --title "$title" --body-file "$body_file")
    if gh label list --json name --jq '.[].name' 2>/dev/null | grep -Fxq "changelog/feature"; then
      pr_args+=(--label "changelog/feature")
    fi
    pr_url="$(gh pr create "${pr_args[@]}" 2>&1)" || { rm -f "$body_file"; pa_fail "gh pr create failed: $pr_url"; pr_url=""; return 1; }
    jlog "draft PR for '$name': $pr_url"
  fi
  rm -f "$body_file"

  git checkout -f "$ORIG_BRANCH" >/dev/null 2>&1 || true
  pa_result succeeded "$(jq -n --argjson mr "$material_risk" --arg rs "$risk_summary" --argjson cf "$changed_json" \
    '{material_risk:$mr, risk_summary:$rs, changed_files:$cf}')"
  return 0
}

# build_generate_body — writes the add/update PR body to a temp file, prints path.
build_generate_body() {
  local name="$1" svc="$2" action="$3" test_func="$4" test_f="$5" risk_callout="$6" bullets="$7" warn_md="$8" howto="$9"
  local f; f="$(mktemp -t tfgen-batch-body.XXXXXX.md)"
  local verb_lc="added"; [[ "$action" == "updated" ]] && verb_lc="updated"
  cat >"$f" <<EOF
${DISCLAIMER_TOP}

${DISCLAIMER_VERIFY}

${risk_callout}## ${name} data source (generator-v2)

This data source was ${verb_lc} by tfgen from the annotated Datadog v2 OpenAPI spec, as part of
a batch that keeps the generated set in lockstep with the spec. It is registered in
\`datasources_generated.go\`'s \`generatedDatasources\` slice (tfgen owns that file).

**Spec hash:** \`${SPEC_HASH:-unknown}\`

### Generated
- \`datadog/fwprovider/data_source_datadog_${name}.go\`
- \`datadog/tests/data_source_datadog_${name}_test.go\`
- \`datadog/fwprovider/datasources_generated.go\` — registers the constructor
- \`docs/data-sources/${name}.md\`

### Test / cassette
- Acceptance test: \`${test_func}\` in \`${test_f}\`
- Cassette: scaffold, not yet recorded

### Reviewer notes / risks
${bullets}
${warn_md}

### How to test
${howto}

${DISCLAIMER_FOOT}
EOF
  printf '%s' "$f"
}

# build_retire_body — writes the retirement PR body to a temp file, prints path.
build_retire_body() {
  local name="$1"
  local f; f="$(mktemp -t tfgen-batch-body.XXXXXX.md)"
  cat >"$f" <<EOF
${DISCLAIMER_TOP}

## Retire ${name} data source (generator-v2)

The \`x-datadog-tf-generator\` annotation for \`datadog_${name}\` is no longer present in the
spec, so tfgen retired it: this PR deletes the generated data source, its test scaffold and its
docs page, and removes the constructor from \`datasources_generated.go\`. No recorded cassette
was found for it, so it was never verified toward release and is safe to remove.

**Spec hash:** \`${SPEC_HASH:-unknown}\`

### Removed
- \`datadog/fwprovider/data_source_datadog_${name}.go\`
- \`datadog/tests/data_source_datadog_${name}_test.go\`
- \`datadog/fwprovider/datasources_generated.go\` — constructor removed
- \`docs/data-sources/${name}.md\`

> ⚠️ **Reviewer:** if this data source ever *replaced* a hand-written one (\`overwrites:\`), this
> PR does not restore the original — resurrecting it is out of scope and needs a human.

${DISCLAIMER_FOOT}
EOF
  printf '%s' "$f"
}

STAGE="fan-out"
ART_RESULTS=()
run_one() {  # <action> <name>
  local name="$2"
  [[ -z "$name" ]] && return 0
  local out; out="$(process_artifact "$1" "$name")" || true
  [[ -z "$out" ]] && out="$(jq -n --arg n "$name" --arg a "$1" '{artifact_name:$n, action:$a, status:"failed", branch:null, pr_url:null, error:"process_artifact produced no result"}')"
  ART_RESULTS+=("$out")
}
for n in "${CREATED[@]}"; do run_one created "$n"; done
for n in "${UPDATED[@]}"; do run_one updated "$n"; done
for n in "${RETIRED[@]}"; do run_one retired "$n"; done

# ---------------------------------------------------------------------------
# Stage: blocked — a tracking issue for adopted orphans we refused to delete
# ---------------------------------------------------------------------------
STAGE="blocked"
ISSUE_URL=""
if [[ "$N_BLOCKED" -gt 0 && "$NO_PR" -eq 0 ]]; then
  issue_body="$(mktemp -t tfgen-batch-issue.XXXXXX.md)"
  {
    printf 'Reconcile found generated data sources whose annotation is gone but that carry a\n'
    printf 'recorded acceptance-test cassette, so they were adopted toward release. Deleting a\n'
    printf 'published data source is a breaking change, so the batch left them in place. Each\n'
    printf 'needs a proper deprecation cycle by a human:\n\n'
    for n in "${BLOCKED[@]}"; do [[ -n "$n" ]] && printf -- '- [ ] `datadog_%s`\n' "$n"; done
    printf '\n_Filed automatically by generate_batch.sh (spec hash `%s`)._\n' "${SPEC_HASH:-unknown}"
  } >"$issue_body"
  if ISSUE_URL="$(gh issue create --title "Retire adopted generated data sources ($N_BLOCKED)" --body-file "$issue_body" 2>&1)"; then
    jlog "tracking issue: $ISSUE_URL"
  else
    jlog "WARNING: gh issue create failed: $ISSUE_URL"
    ISSUE_URL=""
  fi
  rm -f "$issue_body"
fi

# ---------------------------------------------------------------------------
# Result
# ---------------------------------------------------------------------------
STAGE="done"
ARTIFACTS_JSON="$(printf '%s\n' "${ART_RESULTS[@]:-}" | jq -s '[.[] | select(.!=null and .!="")]' 2>/dev/null || echo '[]')"
FAILED_ARTS="$(printf '%s' "$ARTIFACTS_JSON" | jq '[.[] | select(.status=="failed")] | length')"
OVERALL=succeeded; [[ "$FAILED_ARTS" -gt 0 ]] && OVERALL=partial

RESULT_JSON="$(jq -n \
  --arg status "$OVERALL" --arg base "$BASE" --arg spec "$SPEC" --arg spec_hash "$SPEC_HASH" \
  --argjson counts "$(counts_json)" \
  --argjson artifacts "$ARTIFACTS_JSON" \
  --argjson blocked "$(blocked_json)" \
  --arg issue_url "$ISSUE_URL" \
  --argjson metrics "$(metrics_json)" \
  '{status:$status, dry_run:false, base:$base, spec:{path:$spec, hash:$spec_hash}, counts:$counts,
    artifacts:$artifacts, retire_blocked:$blocked,
    tracking_issue_url:(if $issue_url=="" then null else $issue_url end), metrics:$metrics}')"

jlog "$(printf '%s' "$(metrics_json)" | jq -r '"done in \(.runtime_seconds)s | claude: \(.claude_calls) calls, $\(.claude_cost_usd)"')"
emit_result "$RESULT_JSON"
