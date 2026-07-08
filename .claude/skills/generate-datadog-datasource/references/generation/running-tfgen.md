# Phase 2b — Run tfgen, gate, and commit

Input: the slice path from Phase 2a (`slice-and-annotate.md`) and the known scenario from
Phase 1. Output: the generated files committed on a fresh branch, plus the RunReport and
branch name carried into Phase 3. Run everything from the **repo root** (tfgen's output
paths are relative to cwd). Use `make` targets, never raw `go`.

## 1. Ensure the tfgen binary exists

```bash
ls bin/tfgen 2>/dev/null || make tfgen-build   # builds bin/tfgen
```

## 2. Generate, capturing the RunReport

```bash
./bin/tfgen generate --spec "$SLICE" --emit-tests --report /tmp/tfgen-report.json
```

- `--emit-tests` also scaffolds the acceptance test (`datadog/tests/data_source_datadog_<name>_test.go`); without it you get no test file.
- `--report <path>` writes the RunReport JSON where you can parse it (`-` = stdout).
- Default output roots: `--output-root datadog/fwprovider`, `--tests-output-root datadog/tests`.
- Exit codes: `0` ok · `3` `--check` found changes (not used here) · nonzero = failure.

The RunReport shape: `artifacts[]` (`{name, status, path}`), `summary` (counts incl.
`failed`), `diagnostics[]` (`{severity, message, location}`).

## 3. Gate on the report — before committing anything

**Stop** and do not commit if either is true:
- `summary.failed > 0`, or
- any `diagnostics[].severity == "error"`.

When gated: quote the failing artifact and each error diagnostic's `message`/`location`,
state the concrete fix, and leave the working tree **uncommitted** for inspection. That
explanation is the deliverable for the run — nothing is pushed. To reset a bad run:

```bash
git checkout -- datadog/ && git clean -fd datadog/fwprovider datadog/tests docs/data-sources
```

`warning`/`info` diagnostics do **not** gate — carry them into the Phase 3 risk section.

## 4. Docs and build

```bash
make docs     # generates docs/data-sources/<name>.md via tfplugindocs
make build    # compiles the provider; fails loudly if the generated code doesn't build
```

If `make docs` produces no file for this data source, the registration didn't take — see
troubleshooting below. If `make build` fails, treat it like a gate: fix or report, don't
commit broken code.

## 5. What tfgen changed — the files to commit

Check `git status`. Expect:

| File | When |
|---|---|
| `datadog/fwprovider/data_source_datadog_<name>.go` | always (the data source) |
| `datadog/tests/data_source_datadog_<name>_test.go` | always (`--emit-tests`) |
| `datadog/fwprovider/datasources_generated.go` | always — tfgen rewrites the `generatedDatasources` slice to register the new constructor |
| `docs/data-sources/<name>.md` | after `make docs` |
| `datadog/fwprovider/framework_provider.go` | **only when overwriting** — the retired hand-written constructor is removed from `Datasources` |

Do **not** commit the temp slice (`$TMPDIR/tfgen-slices/*.yaml`) — it's a transient input.
Do **not** hand-edit `datasources_generated.go`; tfgen owns it.

## 6. Branch off master and commit

This phase owns branching — tfgen does no git.

```bash
# only branch if not already on a feature branch
[ "$(git rev-parse --abbrev-ref HEAD)" = master ] && \
  git checkout -b generate/datadog_<name>_datasource

git add datadog/fwprovider/data_source_datadog_<name>.go \
        datadog/tests/data_source_datadog_<name>_test.go \
        datadog/fwprovider/datasources_generated.go \
        docs/data-sources/<name>.md
# add framework_provider.go too if this was an overwrite
git commit -m "[<service>] Add datadog_<name> data source (generated)"
```

Confirm the branch name with the user (the example `generate/datadog_<name>_datasource` is a
default). Carry into Phase 3: the RunReport (`/tmp/tfgen-report.json`), the known
scenario/cardinality, the slice path, and the branch name. Do not open the PR here.

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `overwrites target %q not found in the framework Datasources slice` | The `--overwrites` constructor isn't a hand-written framework data source (e.g. it's an SDKv2 `DataSourcesMap` entry, or a typo). Fix the target or drop `--overwrites`. |
| `make docs` shows no new file | Registration didn't take — confirm the constructor is in `datasources_generated.go`'s `generatedDatasources` slice (tfgen writes it). |
| `make build` fails on the generated file | Report it as a gate failure with the compiler output; don't commit. |
| Report has `warning`/`info` only | Not a gate — commit, and carry the diagnostics into the Phase 3 risk section. |
