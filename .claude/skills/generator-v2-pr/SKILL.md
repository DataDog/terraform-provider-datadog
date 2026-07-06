---
name: generator-v2-pr
description: >
  Open a review-ready GitHub PR for a Terraform data source that tfgen (generator-v2)
  has already generated and committed onto the current branch. The skill parses the
  tfgen RunReport JSON, classifies the data source scenario from its
  x-datadog-tf-generator annotation, evaluates it against the matching emit golden and
  a set of runtime-risk heuristics, pulls build/vet/fmt status from GitHub Actions
  checks, drafts a standard PR body (what was generated, cardinality semantics, prominent
  risk callouts, and a manual record/replay testing guide), and opens the PR with `gh`.
  Use this skill whenever a generated Datadog data source needs a PR, or when the user
  mentions tfgen, generator-v2, a RunReport / report JSON, opening a PR for a generated
  data source, evaluating a generated data source against goldens, checking a generated
  data source for merge risks, or writing cassette / acceptance-test instructions — even
  if they don't say the word "skill".
compatibility: Requires `git`, `gh` (authenticated), and a checkout of the terraform-provider-datadog repo with the generated files already committed on the current branch.
---

# generator-v2 PR flow

tfgen deterministically generates Terraform provider **data source** code (singular,
plural, or both) plus acceptance tests from an annotated OpenAPI spec. By the time this
skill runs, tfgen has **already created the branch and committed the files**. This skill
turns that state into a review-ready GitHub PR.

The PR must do four things, in this order of importance:
1. Say plainly **what was generated**.
2. **Flag every merge risk**, however small — reasoned from what the endpoint actually does.
3. Explain **how the generated source may differ** from the ideal shape / handwritten conventions.
4. Give a **complete manual testing guide** (cassettes are recorded and replayed by hand).

## The one rule that overrides everything: never overclaim "verified"

A green RunReport and a clean build prove only that code was **generated and compiles** —
not that it works at runtime. A plural data source has built and reported `created`
cleanly yet returned 0 rows live (read-after-write lag / silent-empty trap; see
`references/risk-heuristics.md`). So:

- Only use the words "verified", "working", "confirmed", or "replays green" if a **cassette actually replayed green**.
- If generation/build succeeded but no cassette has been replayed, say exactly that: "generated and builds cleanly; runtime behavior not yet verified — see testing guide."
- When in doubt, describe what was checked, not what you assume.

This is the single most common way to write a misleading PR here. Guard against it in every section you draft.

## Inputs and where to find them

| Signal | Source | Notes |
|---|---|---|
| What was generated | tfgen **RunReport JSON** | `artifacts[]`, `summary`, `diagnostics[]`. Default emit is stdout (`--report -`). See below for locating it. |
| Scenario / cardinality | `x-datadog-tf-generator` annotation on the backing GET op | Drives which risk heuristics apply. Report JSON does **not** carry cardinality. |
| Ideal shape | emit golden for that scenario | `.generator-v2/internal/testdata/emit/*.golden` — scenario fixtures, **not** per-data-source references. |
| Generated code | `datadog/fwprovider/data_source_datadog_<name>.go` and `datadog/tests/data_source_datadog_<name>_test.go` | Same `Name` in the report, two entries differing by `Path`. |
| Build / vet / fmt | **GitHub Actions job conclusions** (via `gh`) | There is **no build-log artifact file**. Raw stdout goes to the Actions log. Query it with `gh` after the PR exists. |

### Locating the RunReport
The report is emitted to `--report` (default stdout, `-`), so it may not be sitting on
disk after tfgen ran. Prefer, in order:
1. A committed/emitted report file if one exists in the working tree.
2. Re-run to regenerate it (idempotent):
   ```bash
   cd .generator-v2 && ./bin/tfgen generate --report -
   ```
   Exit codes: `0` ok · `3` `--check` found changes · nonzero = failure.

If you cannot obtain a report, stop and tell the user — do not fabricate the "Generated" list from the git diff alone.

## Workflow

### 1. Preconditions
- Confirm `gh auth status` is authenticated. The PR opens as whoever is logged into `gh`; there is no bot/token.
- Confirm HEAD is **not** `master`. If it is, refuse and explain — tfgen generates onto the current branch and the PR head must be that branch.
- Obtain the RunReport (above).

### 2. Gate on the report — before doing anything else
Parse the report and **stop** (do not open a PR) if either is true:
- `summary.failed > 0`, or
- any `diagnostics[].severity == "error"`.

When gated, write a short explanation of exactly which artifact failed and which diagnostics fired, quote the diagnostic `Message`/`Location`, and state what intervention is needed. That explanation is the deliverable for this run — nothing gets pushed.

`warning`/`info` diagnostics do **not** gate; carry them into the risk section.

### 3. Classify the scenario
Read the `x-datadog-tf-generator` annotation for the operation backing this data source to determine the scenario. The matrix:

| Scenario | Annotation | Cardinality |
|---|---|---|
| by-id only | `group{ read: GetX }` | singular |
| both (id-optional) | `group{ read: GetX, search: ListX }` | (none) |
| search-only | `group{ search: ListX }` | singular |
| plural | `group{ read: ListX }` + `cardinality: plural` | plural |

If you can't find the annotation, infer from the generated `.go`: a computed-only `id` with a list read → plural; a required `id` → by-id; presence of a search block → both/search-only. Note in the PR that cardinality was inferred, not read from the annotation.

### 4. Read the ideal shape
Open the matching emit golden(s) under `.generator-v2/internal/testdata/emit/` — e.g.
`plural.golden` / `plural_no_params.golden` / `plural_nested.golden` for plurals,
`singular_both.golden` / `singular_search.golden` etc. for singulars, plus the
corresponding `data_source_test_{singular,singular_search,plural}.golden`. These are the
**intended templates for that scenario**, not a reference for this specific endpoint. Use
them to understand what "correct for this shape" looks like — then judge fit against the
actual endpoint.

### 5. Evaluate — this is the core, and it requires reasoning, not diffing
Read the generated source and test, then read the spec for this endpoint (fields, enums,
required fields, what the **list** payload contains vs. the **detail** payload,
pagination markers). Work through `references/risk-heuristics.md` and, more importantly,
reason about what this endpoint actually does and whether the scenario default fits it.

The generator is well-built, so genuine defects usually surface as build errors or report
failures (already gated in step 2). What remains are **logical** risks — places where the
generated code is valid but the scenario default is a poor fit for this endpoint's
semantics. Those are exactly what the reviewer needs flagged. Do not pad the risk section
with boilerplate; flag what genuinely applies, and say why, tied to this endpoint.

### 6. Build / vet / fmt status
There is no artifact file to parse. After the PR exists (step 9), read the checks:
```bash
gh pr checks "$(git rev-parse --abbrev-ref HEAD)"
# for any failing check:
gh run view <run-id> --log-failed
```
Relevant workflows: `tfgen.yml` ("Generator Checks": build + `go vet ./cmd/tfgen/...`,
working-dir `.generator-v2`) and `test.yml` (pre-commit, golangci-lint, `make vet`,
`make license-check`, `make check-docs`, `make testall`/`make testacc`). If a check fails,
summarize the failing step's output and state the concrete fix needed. If checks are still
pending, say "checks pending" — don't guess a conclusion.

### 7. Draft the PR body
Use the template in `references/pr-body-template.md` **exactly**, including the footer line
verbatim. Populate "Generated" from `artifacts[].{name,status,path}`. The body **must**
open with the **project-context disclaimer** (what this auto-generation project is; PR
description generated and verified via AI; review and test before using), followed
immediately by the auto-generated **verification disclaimer** — both are required and
distinct from the footer. If step 5 found any **material** risk, additionally prepend the
prominent risk callout directly under the disclaimers so risks aren't buried.

**Docs check.** `docs/data-sources/<name>.md` comes from `make docs` (tfplugindocs), not
tfgen. Check the branch: if the docs file is present, list it in `### Generated` as
`(created)`. If it's absent, drop that line and add the big-text "Run `make docs` before
merge" block (see template) — it's an easy, prominent action item, and CI's docs check
fails without the docs file.

### 8. CI-required PR metadata — the PR will NOT pass CI without all three
These are enforced at the repository level; get them wrong and CI never goes green:
1. **Verification disclaimer** in the body (step 7 / template) — states the PR contains
   auto-generated code that must be verified before merge.
2. **Service in brackets in the title.** The title must be
   `[<service>] Add datadog_<name> data source`. The `<service>` prefix is enforced by the
   repo's semantic-title check. Derive `<service>` from the endpoint's spec tag /
   annotation (e.g. `rum` for `rum_applications`); if you can't determine it confidently,
   ask the user rather than guess — a wrong or missing prefix fails CI.
3. **`changelog/feature` label** on the PR.

### 9. Open the PR
```bash
git push -u origin "$(git rev-parse --abbrev-ref HEAD)"
gh pr create --base master --head "$(git rev-parse --abbrev-ref HEAD)" \
  --title "[<service>] Add datadog_<name> data source" \
  --label "changelog/feature" \
  --body-file <path-to-drafted-body>
```
`gh pr create` fails if the `changelog/feature` label doesn't exist in the repo — if that
happens, surface the error rather than dropping the label (the PR needs it to pass CI).
Opening a PR publishes content on the user's behalf — confirm the drafted title, body, and
label with the user before running `gh pr create`.

### 10. Report back
Tell the user the PR URL, the gate result, the risks you flagged (and why), and the build
status. Be precise about verification per the rule at the top.

## References
- `references/risk-heuristics.md` — scenario-specific runtime pitfalls (Q8). Read every run.
- `references/pr-body-template.md` — the exact PR body + footer.
- `references/testing-guide.md` — Frog-org record/replay + cassette instructions used to build the "Test / cassette" section (confirmed commands, env vars, and cassette naming).

## Scope note
This is the crude, locally-runnable version of this skill, in active development. Expect
rough edges. The generator-v2 code is not yet merged to master, which is why generated
data sources are hand-wired into `framework_provider.go`'s `Datasources` slice rather than
via the `datasources_generated.go` registry — reflect that in the PR body until generator-v2
merges. **Do not** put any "demo/crude/in-development" language into an actual PR except the
single footer line defined in the template.
