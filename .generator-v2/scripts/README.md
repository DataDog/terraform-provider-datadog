# `generate_headless.sh` — headless data source generation

Generates a Datadog Terraform data source end to end and opens a **draft** PR, with no
prompts and no human in the loop. It is the non-interactive counterpart to the
`generate-datadog-datasource` Claude skill: the skill asks you questions as it goes; this
script takes every answer up front as a flag and fails fast if anything is unclear.

The script is a deterministic orchestrator. It runs `slice_and_annotate.py` and `tfgen`,
gates on the generator report, runs `make docs` / `make build`, checks only the expected
files changed, then branches, commits, and opens the PR. The only non-deterministic step is
two small calls to the `claude` CLI that write the PR's risk notes and "How to test"
section; both degrade gracefully if `claude` is missing.

## What it does, in order

1. Validate the arguments and the environment; stop if the working tree is dirty.
2. Start a fresh branch from `--base` (see below).
3. Download the v2 OpenAPI spec (latest by default) and record which commit it was.
4. Slice + annotate the chosen operation(s) with `slice_and_annotate.py`.
5. Run `tfgen`, then stop if the report has any failure or error.
6. `make docs` and `make build`; stop if either fails.
7. Confirm only the generated files changed.
8. Ask `claude` for the risk notes and testing steps (skipped cleanly if unavailable).
9. Commit, push, and open a **draft** PR against `--base`.

On any failure it stops, prints a JSON error, and does not commit or push. It never edits
generated code to "fix" a failure — a failed run's report is the deliverable.

## Prerequisites

On `PATH`: `git`, `gh` (authenticated), `python3` with PyYAML, `make`, `curl`, `jq`.

For the generation itself, the same toolchain `make docs` / `make build` already need:
Go, Terraform, `tfplugindocs`, and `goimports`.

Optional but recommended: the `claude` CLI, authenticated (or `ANTHROPIC_API_KEY` set). If
it is absent, the PR still opens but its risk notes say a reviewer must scan manually and
the testing steps fall back to a fixed block.

## Usage

```
.generator-v2/scripts/generate_headless.sh \
  --artifact-name NAME --cardinality {singular|plural} (--read OP | --search OP) [flags]
```

Run it from anywhere inside the provider checkout — it finds the repo root itself.

### Required

| Flag | Meaning |
|---|---|
| `--artifact-name NAME` | TF name without the `datadog_` prefix. `^[a-z][a-z0-9_]*$`, ≤64 chars, and it may not end in `_test` (that would make the generated file a Go test file). |
| `--cardinality VALUE` | `singular` or `plural`. Always explicit — never guessed. |
| `--read OP` / `--search OP` | operationId(s); at least one. For `plural`, put the list/collection GET in `--read`. |

### Optional

| Flag | Default | Meaning |
|---|---|---|
| `--tf-description TEXT` | derived from the name | Doc string for the data source. |
| `--overwrites CTOR` | none | Retire a hand-written constructor (e.g. `NewDatadogTeamDataSource`) in place. Required if a hand-written data source with this name already exists. |
| `--service NAME` | derived from the spec tag | The `[service]` prefix CI requires in the PR title. |
| `--spec PATH` | curl upstream | Use a local spec file instead of downloading. |
| `--spec-ref REF` | `master` | Which ref of the upstream spec repo to download. |
| `--base BRANCH` | current branch | The branch the PR targets **and** is built from, so the PR shows only the generated files. Defaults to the branch you run the script on. |
| `--branch NAME` | `generate/datadog_<name>_datasource` | The feature branch to create. |
| `--no-pr` | off | Stop after the local commit — no push, no PR. Use this for a safe dry run. |
| `--output-json PATH` | none | Also write the result JSON to this file. |

## Examples

```bash
# Singular, read-only:
.generator-v2/scripts/generate_headless.sh \
  --artifact-name incident_type --cardinality singular --read GetIncidentType

# Singular, id-optional (by-id read + list search):
.generator-v2/scripts/generate_headless.sh \
  --artifact-name team --cardinality singular --read GetTeam --search ListTeams

# Plural (the collection GET is the read):
.generator-v2/scripts/generate_headless.sh \
  --artifact-name teams --cardinality plural --read ListTeams

# Target master explicitly (default is the branch you're currently on):
.generator-v2/scripts/generate_headless.sh \
  --artifact-name teams --cardinality plural --read ListTeams --base master

# Dry run — generate and commit locally, open nothing:
.generator-v2/scripts/generate_headless.sh \
  --artifact-name teams --cardinality plural --read ListTeams --no-pr
```

## Output

The result JSON goes to **stdout**; all human logs go to **stderr**. Exit code `0` means a
draft PR was opened (or, with `--no-pr`, a local commit was made); any other code means it
stopped, and the JSON says where and why.

```bash
.generator-v2/scripts/generate_headless.sh ... > result.json 2> run.log
jq '.status, .pr_url, .metrics' result.json
```

Every run — success or failure — includes a `metrics` block:

```json
"metrics": {
  "runtime_seconds": 137,
  "claude_cost_usd": 0.021,
  "claude_input_tokens": 42500,
  "claude_output_tokens": 860,
  "claude_calls": 2
}
```

`runtime_seconds` is the whole script's wall-clock time; the `claude_*` fields sum both
prose calls (and count cost even when a reply was unusable, since it was still billed).

## Safety

- **Fail-fast.** Ambiguous input, a bad operationId, a failing report, a build failure, or
  anything outside the expected file set stops the run — it does not guess or patch.
- **Draft PR only, always unverified.** The PR is always a draft and always carries the
  "must be verified before merging" disclaimer. Recording the acceptance test against the
  Frog org and replaying it green is a separate, human step; the script never claims the
  data source is verified.
- **Overwrite is opt-in.** It will not retire a hand-written data source unless you pass
  `--overwrites`.

## Running it in a pipeline

The script assumes no particular CI system — a pipeline just needs to give a fresh runner
what a laptop already has:

1. Check out the repo with a git token that can push and open PRs. Use a real PAT or app
   token, not the default CI token, or the PR will not trigger the repo's own CI checks.
2. Install the prerequisites above, including the `claude` CLI.
3. Provide two secrets as environment variables: `ANTHROPIC_API_KEY` for `claude`, and the
   git token for `gh`.
4. Run the script with the same flags (typically wired up as pipeline inputs).
5. Read the JSON from stdout — use `.status` and the exit code as the pass/fail gate, and
   `.pr_url` and `.metrics` for reporting runtime and token cost.

---

# `generate_batch.sh` — keep the provider in lockstep with an annotated spec

Where `generate_headless.sh` adds **one** data source, `generate_batch.sh` reconciles the
**whole** generated set against an already-annotated spec and fans each impacted artifact out
to its own **draft** PR:

- an annotation appears → the data source is **created**,
- a schema under an annotation changes → it is **updated**,
- an annotation is removed → the data source is **retired** (deleted), unless a recorded
  cassette shows it was adopted — those are flagged, never deleted.

The spec is assumed **already annotated** with `x-datadog-tf-generator`; this script does not
slice or annotate (pass `--spec`). It never needs test credentials — it emits code and opens
draft PRs; a human records cassettes afterward.

## What it does, in order

1. Validate arguments/environment; stop if the working tree is dirty.
2. `make tfgen-build`.
3. **One** expensive run on a staging branch: `tfgen generate --reconcile` (generate-all +
   retire orphans) → `make docs` → `make build`. This proves the whole impacted set compiles
   together. On any failure it restores the tree and stops.
4. Capture the generated docs + report, then restore the tree to base.
5. Enforce `--max-prs`; with `--dry-run`, print the plan and stop here (no branches/PRs).
6. **Fan out** (fail-slow — a bad artifact is recorded and skipped, never aborting the batch):
   for each created / updated / retired artifact, cut a branch from base, re-emit just that
   artifact (`--include`) or retire it (`--retire`), confirm only its files changed, and open
   a draft PR. Build is verified in aggregate (step 3); each PR's own CI is the per-artifact net.
7. For adopted orphans left in place, open one GitHub tracking issue.
8. Emit one aggregate batch JSON.

## Usage

```
.generator-v2/scripts/generate_batch.sh --spec PATH [flags]
```

| Flag | Default | Meaning |
|---|---|---|
| `--spec PATH` | — (required) | Full v2 OAS, already annotated with `x-datadog-tf-generator`. |
| `--base BRANCH` | current branch | Branch each PR targets **and** is built from. |
| `--max-prs N` | `25` | Abort before fan-out if more than N PRs would open. |
| `--dry-run` | off | Verify + print the plan; open no branches or PRs. |
| `--no-ai` | off | Skip all `claude` calls; PR bodies use deterministic notes only. |
| `--no-pr` | off | Commit each artifact on its branch but do not push or open a PR. |
| `--output-json PATH` | none | Also write the final batch JSON here. |

```bash
# See what a batch would do, without touching git or GitHub:
.generator-v2/scripts/generate_batch.sh --spec annotated-openapi.yaml --dry-run | jq .

# Open the PRs, targeting master, capped at 40:
.generator-v2/scripts/generate_batch.sh --spec annotated-openapi.yaml --base master --max-prs 40
```

## Output

Result JSON on **stdout**, human logs on **stderr**. `status` is `succeeded` (all artifacts
handled), `partial` (some artifacts failed but the batch finished), `planned` (`--dry-run`),
or `failed` (a setup/verify gate stopped it). Each artifact carries its own `status`, `branch`,
`pr_url`, and any risk summary; `retire_blocked` lists the adopted orphans, and
`tracking_issue_url` points at the issue filed for them.

## Safety

- **Nothing runs on a dirty tree**, and the staging/fan-out resets only ever discard changes
  this script created (guarded so a precondition failure never touches your work).
- **Build verified in aggregate.** The single `make build` covers the whole set; per-branch
  builds are intentionally skipped, so each PR relies on its own CI build.
- **Retirement is gated.** Only never-adopted (no recorded cassette) tfgen-generated files are
  deleted; adopted ones and any non-generated file are refused and flagged for a human.
- **Draft PRs only, always unverified**, carrying the same "must be verified before merging"
  disclaimer as the single-artifact script.
- **`overwrites` retirement is out of scope.** If a retired data source had replaced a
  hand-written one, the original is not resurrected — the retire PR flags this for a reviewer.

## Cross-PR note

Each PR edits the sorted `datasources_generated.go`; the first to merge shifts the base, so a
later PR may need a trivial rebase. The resolution is deterministic — re-run the scoped emit.
