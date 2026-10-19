---
name: generate-datadog-datasource
description: >
  Generate a Datadog Terraform provider data source end to end and open a review-ready
  GitHub PR for it. Runs in three phases: (1) Input — collect the read group
  (operationIds), artifact name, cardinality, description, and overwrite target; (2)
  Generation — build an annotated OpenAPI slice with slice_and_annotate.py, run tfgen on
  it, run make docs/build, and commit onto a new branch; (3) PR — a quick runtime-risk scan
  (trusting the generator for correctness), draft the standard PR body with disclaimers +
  testing guide, and open the PR with `gh`. Use
  this skill whenever the user wants to generate a Datadog data source, mentions tfgen /
  generator-v2, slice_and_annotate, an OpenAPI operation they want a data source for,
  opening a PR for a generated data source, evaluating generated code against goldens, or
  writing cassette / acceptance-test instructions — even if they don't say "skill".
compatibility: Requires `git`, `gh` (authenticated), Python 3 + PyYAML, a checkout of the terraform-provider-datadog repo, and network access to the full Datadog v2 OpenAPI spec (curled from upstream by default; overridable via `--spec` or `$DATADOG_OPENAPI_V2_SPEC`).
---

# Generate a Datadog data source (input → generation → PR)

This skill takes a Datadog v2 OpenAPI operation from nothing to a review-ready PR. It owns
the **whole** flow — the earlier version assumed tfgen had already generated and committed;
this version runs generation itself, so it also owns branching and committing.

```
Phase 1: INPUT        Phase 2: GENERATION                    Phase 3: PR
collect params  ──▶   slice_and_annotate.py → slice     ──▶  classify scenario
(routes, name,        tfgen generate → .go + test             evaluate vs goldens + risks
 cardinality,         make docs / make build                  draft PR body
 description,         branch off master + commit               open PR with gh
 overwrite)           (gate on the RunReport first)            report back
```

Each phase has its own reference subdirectory under `references/`. Read the reference for a
phase before running it.

## Operating principles — run it fast and hands-off

tfgen is deterministic and well-tested. This skill is a **thin wrapper** around it, not an
audit. Three rules keep runs fast and honest:

1. **Trust the generator.** If tfgen produced the files and `make build` passed, treat the
   output as correct. Do not re-derive its decisions, re-verify field-by-field, or
   second-guess the scenario it emitted. Report what happened — don't prove the code right.
2. **Never fix — report and stop.** On ANY failure (spec/annotation error, generation
   failure, `make build`/`make docs` failure, or a red CI check) do **not** edit the spec,
   patch the generated code, retry, or work around it. Quote the error verbatim, say plainly
   why the data source could not be generated, and stop. For a failed run, that report *is*
   the deliverable.
3. **Keep analysis light.** The only judgment that matters is a quick scan for *material*
   runtime risks (`references/pr/risk-heuristics.md`) — minutes, not a line-by-line review,
   and not a golden diff. If nothing clearly applies, say so and move on.

## The one rule that overrides everything: never overclaim "verified"

A green RunReport and a clean build prove only that code was **generated and compiles** —
not that it works at runtime. A plural data source has built and reported `created` cleanly
yet returned 0 rows live (read-after-write lag / silent-empty trap; see
`references/pr/risk-heuristics.md`). So:

- Only use "verified", "working", "confirmed", or "replays green" if a **cassette actually replayed green**.
- If generation/build succeeded but no cassette has replayed, say exactly that: "generated and builds cleanly; runtime behavior not yet verified — see testing guide."
- When in doubt, describe what was checked, not what you assume.

This is the single most common way to write a misleading PR here. Guard against it in every section you draft, and never let the confidence of having *generated* the code leak into runtime claims.

---

## Phase 1 — Input

**Goal:** produce a complete, validated parameter set for `slice_and_annotate.py`. Nothing
is generated in this phase — you are only deciding *what* to generate.

Collect, confirming each with the user:
- **Spec path** — default: curl the upstream v2 spec to a temp file (see `references/input/collecting-inputs.md`); an explicit path or `$DATADOG_OPENAPI_V2_SPEC` overrides. A local copy must exist before you can discover routes.
- **Read group (routes)** — the operationIds. If the user names a resource/service and a spec is available, inspect it and propose candidate GET routes; otherwise take operationIds directly. Validate every operationId against the spec.
- **Cardinality + scenario** — singular (by-id / both / search-only) vs plural, derived from which GETs exist and what the user wants.
- **Artifact name** — default derived from the resource (snake_case, no `datadog_` prefix); validate `^[a-z][a-z0-9_]*$`, ≤64.
- **TF description** — default `Use this data source to retrieve information about an existing Datadog <thing>.` (plural: `…existing Datadog <thing>s.`); only ask if they want a custom one.
- **Overwrite** — auto-detect whether a hand-written `datadog_<name>` data source already exists; if so, find its constructor and ask whether to retire it (`--overwrites`). Otherwise additive.

End the phase by echoing the full parameter set back and getting a go-ahead.

**Details:** `references/input/collecting-inputs.md`.

---

## Phase 2 — Generation

**Goal:** turn the parameter set into committed generated files on a fresh branch. Do not
open a PR here.

1. **Preconditions.** `gh auth status` authenticated. If HEAD is `master`, that's fine — this phase creates the branch. Ensure `bin/tfgen` exists (`make tfgen-build` if not).
2. **Build the slice.** Call `slice_and_annotate.py` with the phase-1 params; capture the printed slice path (stdout is only the path). See `references/generation/slice-and-annotate.md`.
3. **Generate.** Run `tfgen generate --spec "$SLICE" --report -`, capturing the RunReport JSON. See `references/generation/running-tfgen.md`.
4. **Gate on the report — before committing.** Stop if `summary.failed > 0` or any `diagnostics[].severity == "error"`. Quote the failing artifact + diagnostics **verbatim**, say plainly why it couldn't be generated, and stop — do **not** edit the spec, retry, or fix anything (principle 2). Leave the working tree uncommitted. That report is the deliverable; nothing is committed. `warning`/`info` do **not** gate — carry them into the PR risk section.
5. **Docs + build.** Run `make docs` (creates `docs/data-sources/<name>.md`) and `make build` to confirm it compiles. Use make targets, never raw `go`. If either fails, quote the output and stop (principle 2) — do not attempt to fix the generated code.
6. **Branch + commit.** Create the branch off `master` and commit the generated `.go`, test, and docs files. Carry forward to Phase 3: the RunReport, the known scenario/cardinality, the slice path, and the branch name.

**Details:** `references/generation/slice-and-annotate.md`, `references/generation/running-tfgen.md`.

---

## Phase 3 — PR

**Goal:** turn the committed branch into a review-ready PR. The scenario and RunReport are
already known from Phase 2 — do not re-derive them; just carry them in.

1. **Quick risk scan.** Skim `references/pr/risk-heuristics.md` and flag only the risks that *clearly* apply to this endpoint (e.g. paginated plural, sensitive detail-only fields, path-nested by-id). This is a fast pass, not a code audit — trust the generator for correctness. If nothing material jumps out, say so and move on.
2. **(Optional) Golden sanity-check.** Only if a specific risk needs confirming, open the matching emit golden under `.generator-v2/internal/testdata/emit/` (the scenario template) to check that one doubt. Otherwise skip — do not diff the generated code against goldens by default.
3. **Draft the PR body.** Use `references/pr/pr-body-template.md` exactly: project-context disclaimer first, test-scaffold disclaimer second, verification disclaimer third, then a prominent risk callout if any material risk was found, then the docs callout if `docs/data-sources/<name>.md` is absent, then the body. Populate "Generated" from `artifacts[].{name,status,path}`.
4. **CI-required metadata (all three, or CI never goes green).** Verification disclaimer in the body; title `[<service>] Add datadog_<name> data source` (derive `<service>` from the spec tag; ask if unsure — a wrong prefix fails CI); `changelog/feature` label.
5. **Open the PR.** Push the branch, then `gh pr create`. Opening a PR publishes on the user's behalf — confirm the drafted title, body, and label with the user first.
6. **Checks + report back.** After the PR exists, read `gh pr checks`; for a failing check, `gh run view <run-id> --log-failed` and report which check failed with its output quoted — do **not** attempt to fix it (principle 2). Report the PR URL, the gate result, the risks flagged (and why), and the build/check status — precise about verification per the top rule.

**Details:** `references/pr/risk-heuristics.md`, `references/pr/pr-body-template.md`, `references/pr/testing-guide.md`.

---

## References
- `references/input/collecting-inputs.md` — Phase 1: the parameter set, spec resolution, route discovery, scenario/cardinality decision, overwrite auto-detect.
- `references/generation/slice-and-annotate.md` — Phase 2: how to call `slice_and_annotate.py` and how the annotation works.
- `references/generation/running-tfgen.md` — Phase 2: build tfgen, generate, gate on the report, make docs/build, branch + commit.
- `references/pr/risk-heuristics.md` — Phase 3: scenario-specific runtime pitfalls. Read every run.
- `references/pr/pr-body-template.md` — Phase 3: the exact PR body + footer.
- `references/pr/testing-guide.md` — Phase 3: Frog-org record/replay + cassette instructions.

## How the generated data source is registered
tfgen owns `datadog/fwprovider/datasources_generated.go`: every run rewrites its
`generatedDatasources` slice from the set of data sources produced. `framework_provider.go`
appends that slice alongside the hand-written `Datasources`, so additive generation is wired
up **without editing `framework_provider.go`**. `framework_provider.go` is edited only in the
**overwrite** case — to remove the retired hand-written constructor from its `Datasources`
slice. Reflect this accurately in the PR body; do not repeat the older "hand-wired into
`framework_provider.go`" phrasing.

## Scope note
This is the locally-runnable version of this skill, in active development. Expect rough
edges. **Do not** put any "demo/crude/in-development" language into an actual PR except the
single footer callout defined in the template.
