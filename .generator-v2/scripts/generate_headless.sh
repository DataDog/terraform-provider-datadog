#!/usr/bin/env bash
# Headless spine for generating a Datadog data source and opening a draft PR.
#
# Deterministic orchestrator: validate -> slice -> generate -> gate -> whitelist
# -> docs -> build -> branch -> commit -> draft PR. There is NO LLM in the loop:
# this mirrors the CI splitter (.github/workflows/tfgen-split.yml), whose PR title
# and body are fully static — "no in-CI Claude" is a project decision. A run's PR
# is byte-shaped like a per-artifact PR the splitter fans out.
#
# Safety model: every fork is an explicit flag or a hard failure. Nothing runs on
# a base branch, every PR is a draft, and the verification disclaimer is always
# present — the worst case is a draft PR a human closes.
#
# --cleanup undoes a local run (delete the feature branch, reset the tree to
# --base) so the whole thing can be demoed over and over.
#
# stdout carries ONLY the final JSON result. All human logs go to stderr.
# Exit 0 = success; nonzero = failure (with a JSON {status:"failed",...} on stdout).

set -euo pipefail

# Remove this run's scratch files on any exit. REPORT and the PR-body file are
# excluded — their paths are returned in the result so the output can be inspected
# after the run finishes; only a spec we curled is removed, never a --spec file.
cleanup() {
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
    git clean -fdq datadog/fwprovider datadog/tests docs/data-sources examples/data-sources >/dev/null 2>&1 || true
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
   or: generate_headless.sh --cleanup (--artifact-name NAME | --branch NAME) [--base BRANCH]

Required (generate):
  --artifact-name NAME     snake_case, no datadog_ prefix (^[a-z][a-z0-9_]*$, <=64)
  --cardinality VALUE      singular | plural  (explicit — never inferred)
  --read OP / --search OP  operationId(s); >=1 required. Plural: list op in --read.

Optional:
  --tf-description TEXT     doc string (default derived from the name)
  --overwrites CTOR        retire a hand-written constructor, e.g. NewDatadogTeamDataSource
  --service NAME           informational only (default: derived from the op's spec tag)
  --spec PATH              full v2 OAS file (default: curl upstream)
  --spec-ref REF           git ref of datadog-api-client-go to curl (default: master)
  --base BRANCH            branch the PR targets and is built from (default: current branch)
  --branch NAME            feature branch (default: generate/datadog_<name>_datasource)
  --no-pr                  stop after commit; do not push or open a PR (use this for demos)
  --output-json PATH       also write the final result JSON here

Cleanup:
  --cleanup                undo a local run: switch to --base, delete the feature
                           branch, and discard any stray generated files. Local
                           only — it never touches origin or any PR. Requires
                           --artifact-name or --branch; --base defaults to master.
EOF
  exit 2
}

# metrics_json — this run's wall-clock time as one JSON object; safe anytime.
metrics_json() {
  jq -n --argjson rt "${SECONDS:-0}" '{runtime_seconds:$rt}'
}

# ---------------------------------------------------------------------------
# Args
# ---------------------------------------------------------------------------
ARTIFACT_NAME="" CARDINALITY="" READ_OP="" SEARCH_OP="" TF_DESCRIPTION=""
OVERWRITES="" SERVICE="" SPEC="" SPEC_REF="master" BASE="" BRANCH=""
NO_PR=0 OUTPUT_JSON="" CLEANUP=0

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
    --cleanup)       CLEANUP=1; shift ;;
    --output-json)   OUTPUT_JSON="${2:-}"; shift 2 ;;
    -h|--help)       usage ;;
    *) echo "unknown flag: $1" >&2; usage ;;
  esac
done

# ---------------------------------------------------------------------------
# Stage: cleanup — undo a local run, then stop. Dispatched before the generate
# preconditions so it needs neither the read/cardinality args nor gh/python.
# ---------------------------------------------------------------------------
if [ "$CLEANUP" = 1 ]; then
  STAGE="cleanup"
  for tool in git jq; do
    command -v "$tool" >/dev/null 2>&1 || die "required tool not found: $tool"
  done
  ROOT="$(git rev-parse --show-toplevel 2>/dev/null)" || die "not inside a git checkout"
  cd "$ROOT"

  [ -n "$ARTIFACT_NAME" ] || [ -n "$BRANCH" ] || die "cleanup needs --artifact-name or --branch"
  if [ -n "$ARTIFACT_NAME" ]; then
    [[ "$ARTIFACT_NAME" =~ ^[a-z][a-z0-9_]*$ ]] || die "invalid --artifact-name '$ARTIFACT_NAME'"
  fi
  TARGET="${BRANCH:-generate/datadog_${ARTIFACT_NAME}_datasource}"
  RETURN="${BASE:-master}"
  git rev-parse --verify --quiet "refs/heads/$RETURN" >/dev/null 2>&1 \
    || die "cleanup target base '$RETURN' does not exist locally (pass --base)"

  CUR="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo HEAD)"
  if [ "$CUR" = "$TARGET" ]; then
    git checkout -f "$RETURN" >&2 || die "could not switch off '$TARGET' to '$RETURN'"
    jlog "switched from $TARGET to $RETURN"
  fi

  # Discard any generated files still loose in the working tree (e.g. a run that
  # failed before committing). Scoped to the generator-owned paths.
  git checkout -- datadog docs examples >/dev/null 2>&1 || true
  git clean -fdq datadog/fwprovider datadog/tests docs/data-sources examples/data-sources >/dev/null 2>&1 || true

  REMOVED=false
  if git rev-parse --verify --quiet "refs/heads/$TARGET" >/dev/null; then
    git branch -D "$TARGET" >&2 && REMOVED=true || die "could not delete branch '$TARGET'"
    jlog "deleted branch $TARGET"
  else
    jlog "no local branch '$TARGET' to delete"
  fi

  emit_result "$(jq -n --arg status cleaned --arg branch "$TARGET" --arg base "$RETURN" \
    --arg artifact_name "$ARTIFACT_NAME" --argjson branch_removed "$REMOVED" \
    '{status:$status, branch:$branch, base:$base, branch_removed:$branch_removed, artifact_name:$artifact_name}')"
  exit 0
fi

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
  die "branch already exists locally: $BRANCH (run with --cleanup to remove it)"
fi
if [ "$NO_PR" -eq 0 ] && git ls-remote --exit-code --heads origin "$BRANCH" >/dev/null 2>&1; then
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

# Derive the service tag for the informational [prefix] if not supplied. Unlike the
# old title convention, the title now uses the splitter's static [generated] prefix,
# so a missing service is non-fatal — it is only reported.
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
[ -n "$SERVICE" ] && jlog "service (informational): [$SERVICE]" || jlog "service tag not derived (informational only)"

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
  git clean -fdq datadog/fwprovider datadog/tests docs/data-sources examples/data-sources 2>/dev/null || true
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
EXAMPLE_FILE="examples/data-sources/datadog_${ARTIFACT_NAME}/data-source.tf"
[ -f "$EXAMPLE_FILE" ] || die "tfgen produced no $EXAMPLE_FILE"
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
  "datadog/tests/provider_test.go"
  "$EXAMPLE_FILE"
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
done < <(git status --porcelain --untracked-files=all)

if [ "${#UNEXPECTED[@]}" -gt 0 ]; then
  jlog "UNEXPECTED changed files: ${UNEXPECTED[*]}"
  die "files changed outside the generated-artifact whitelist: ${UNEXPECTED[*]}"
fi
jlog "changed-file whitelist clean"

CHANGED_JSON="$(git status --porcelain --untracked-files=all | sed 's/^...//' | jq -R . | jq -s .)"

# ---------------------------------------------------------------------------
# Stage: PR body — templated deterministically to mirror the CI splitter's
# static, no-AI body (.github/workflows/tfgen-split.yml, non-retired path).
# ---------------------------------------------------------------------------
STAGE="pr-body"

# created vs updated drives the title verb and the body's status word, matching
# the splitter (updated only when an overwrite retired a hand-written source).
if [ -n "$OVERWRITES" ]; then VERB="Update"; STATUS="updated"; else VERB="Add"; STATUS="created"; fi
TITLE="[generated] $VERB datadog_${ARTIFACT_NAME} data source"

# Acceptance-test scaffold identity, referenced in the body's "How to test".
TEST_FILE="datadog/tests/data_source_datadog_${ARTIFACT_NAME}_test.go"
TEST_FUNC="$(grep -oE 'func TestAcc[A-Za-z0-9_]+' "$TEST_FILE" 2>/dev/null | head -1 | sed 's/^func //' || true)"
if [ -z "$TEST_FUNC" ]; then
  # Fall back to the PascalCase name the splitter derives from the artifact name.
  pascal=""; IFS='_' read -ra parts <<<"$ARTIFACT_NAME"
  for p in "${parts[@]}"; do pascal+="$(tr '[:lower:]' '[:upper:]' <<<"${p:0:1}")${p:1}"; done
  TEST_FUNC="TestAccDatadog${pascal}DataSource"
fi

# Example-warning callout: tfgen flags a placeholder example with a warning whose
# message starts "generated example for". The splitter turns those into a
# do-not-merge-as-is callout; reproduce it from our own RunReport.
EXAMPLE_WARN_MD="$(jq -r '[.artifacts[]?.diagnostics[]?
  | select(.severity=="warning" and (.message | startswith("generated example for")))
  | "> - " + .message] | .[]' "$REPORT" 2>/dev/null || true)"
EXAMPLE_CALLOUT=""
if [ -n "$EXAMPLE_WARN_MD" ]; then
  EXAMPLE_CALLOUT="
> 🚨 **Manual \`Example Usage\` required — do not merge as-is.** tfgen could not render a
> complete example for this data source. The committed \`data-source.tf\` (listed below) is a
> placeholder; replace it with a real, working example before merging:
${EXAMPLE_WARN_MD}
"
fi

# Generated-files list: every changed file except the doc, then the doc line —
# same shape the splitter prints from its per-artifact file set.
FILES_MD="$(git status --porcelain --untracked-files=all | sed -e 's/^...//' -e 's/^"//' -e 's/"$//' \
  | grep -vxF "$DOCS_FILE" | sed -e 's/^/- `/' -e 's/$/`/')"
GEN_LIST="${FILES_MD}
- \`${DOCS_FILE}\` (generated via \`make docs\`)"

PR_BODY_FILE="$(mktemp -t tfgen-pr-body.XXXXXX.md)"
cat >"$PR_BODY_FILE" <<EOF
> ℹ️ **This PR is part of a project that auto-generates Terraform provider data
> sources to increase coverage.** The code is generated **deterministically by tfgen from an
> annotated OpenAPI spec, without the use of LLMs**, and **reviewed via AI, though human review
> is still necessary**. If you use this data source, **review it thoroughly and test it** first.

> ⚠️ **This PR contains auto-generated code and must be verified before merging.** Do not
> merge until an acceptance test is recorded and replays green against the Frog org. A clean
> build and a successful generation report do **not** prove runtime correctness.
${EXAMPLE_CALLOUT}
## ${ARTIFACT_NAME} data source (generator-v2)

This data source was ${STATUS} by tfgen from an annotated slice of the Datadog v2 OpenAPI spec.
The generation, documentation, and compile/format gate completed successfully.

### Generated
${GEN_LIST}

### Test / cassette
- Acceptance-test scaffold: \`${TEST_FUNC}\` in \`${TEST_FILE}\`; complete its TODOs before recording.
- Cassette: not recorded.

### Reviewer notes / risks
- Automated risk scan skipped; a reviewer must inspect runtime behavior manually.
- Generated acceptance tests are boilerplate and do not prove runtime correctness.

### How to test
Add and complete the acceptance test, record it once against the Frog org, then replay it:

\`\`\`bash
eval "\$(dd-auth --domain frog.datadoghq.com --force-app-key --no-cache --output)"
export DD_TEST_CLIENT_API_KEY="\$DD_API_KEY" DD_TEST_CLIENT_APP_KEY="\$DD_APP_KEY"

RECORD=true TESTARGS='-run ${TEST_FUNC}' make testacc
RECORD=false TESTARGS='-run ${TEST_FUNC}' make testacc
\`\`\`

> **Merge note:** auto-generated. If the merge queue reports a conflict in
> \`datasources_generated.go\`, resolve by keeping both entries and re-queue
> (append-only registrations).
>
> If this PR updates \`framework_provider.go\`, resolve conflicts manually:
> preserve unrelated upstream edits while keeping this artifact's retired
> hand-written constructor removed.

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
        "datadog/tests/provider_test.go" \
        "$EXAMPLE_FILE" \
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
  jlog "--no-pr set; stopping after commit (PR body at $PR_BODY_FILE)"
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
  --arg status_word "$STATUS" \
  --arg title "$TITLE" \
  --arg service "$SERVICE" \
  --arg read_op "$READ_OP" --arg search_op "$SEARCH_OP" \
  --arg overwrites "$OVERWRITES" \
  --arg branch "$BRANCH" --arg base "$BASE" \
  --arg spec_source "$SPEC_SOURCE" --arg spec_sha "$SPEC_SHA" --arg spec_hash "$SPEC_HASH" \
  --arg report "$REPORT" --arg pr_url "$PR_URL" --arg pr_body_path "$PR_BODY_FILE" \
  --argjson changed_files "$CHANGED_JSON" \
  --argjson warnings "$WARN_DIAGS" \
  --argjson metrics "$METRICS" \
  '{status:$status, verified:false, artifact_name:$artifact_name, cardinality:$cardinality,
    status_word:$status_word, title:$title, service:$service,
    operations:{read:$read_op, search:$search_op}, overwrites:$overwrites,
    branch:$branch, base:$base, spec:{source:$spec_source, sha:$spec_sha, slice_hash:$spec_hash},
    report_path:$report, pr_url:$pr_url, pr_body_path:$pr_body_path,
    changed_files:$changed_files, generator_warnings:$warnings, metrics:$metrics}')"

jlog "$(printf '%s' "$METRICS" | jq -r '"done in \(.runtime_seconds)s"')"
emit_result "$RESULT_JSON"
