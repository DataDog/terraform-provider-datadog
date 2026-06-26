# tfgen runbook (mini-OAS)

Drive the `tfgen` binary directly — no wrapper script. The input is an OpenAPI
spec carrying an `x-datadog-tf-generator` extension; the generator emits the data
source **and wires it into the provider itself**.

All commands run from the repo root:

```bash
cd /Users/jason.tenczar/projects/terraform-provider-datadog
```

---

## How wiring works (read once)

`tfgen generate` writes three kinds of file under `--output-root` (default
`datadog/fwprovider`):

1. `data_source_datadog_<artifact_name>.go` — the generated data source.
2. `datasources_generated.go` — the generator-owned registry; every run rewrites
   it to the **union** of constructors it has produced.
3. `framework_provider.go` — touched **only** when an artifact declares
   `overwrites`, to retire the hand-written constructor it supersedes.

The `overwrites` field in the extension is the switch:

- **`overwrites: NewFooDataSource`** → the generator removes `NewFooDataSource`
  from the hand-written `Datasources` slice and registers the generated
  constructor in `datasources_generated.go`. Use this when a hand-written data
  source already exists.
- **no `overwrites`** → purely additive: the constructor is just added to
  `datasources_generated.go`. Use this for a brand-new data source.

`FrameworkProvider.DataSources()` appends both slices, so either path is fully
wired — no manual edits, no renaming.

---

## One-time setup

```bash
# 1. Build the binary -> bin/tfgen
make tfgen-build

# 2. Annotate the pristine mini-OAS slices -> gen-test/*.yaml
#    (writes the x-datadog-tf-generator extension, incl. overwrites where a
#     hand-written counterpart exists; see _annotate.py OVERWRITES map)
( cd .generator-v2/internal/testdata/mini-oas/scripts && python3 _annotate.py )
```

After step 2 the annotated specs live in
`.generator-v2/internal/testdata/mini-oas/scripts/gen-test/`. Set a shell var so
the generate commands stay short:

```bash
GEN_TEST=.generator-v2/internal/testdata/mini-oas/scripts/gen-test
```

---

## The extension (what the spec carries)

`_annotate.py` writes this onto the backing GET operation. To hand-author a new
slice, add the same block:

```yaml
x-datadog-tf-generator:
  artifact_kind: data_source        # data_source | resource
  artifact_name: team               # tf name without datadog_ prefix
  tf_description: "Use this data source to retrieve ..."
  group:                            # operationIds backing the artifact
    read: GetTeam                   #   by-id GET
    search: ListTeams               #   list GET (singular resolves one match)
  cardinality: plural               # omit for singular; set on the list variant
  overwrites: NewDatadogTeamDataSource   # omit when purely additive (new)
```

---

## Generate each mini-OAS piece (into the live provider)

Each `gen-test/<name>.yaml` holds exactly one annotated operation, so one
`--spec` produces one data source. `--output-root` defaults to
`datadog/fwprovider`, so these write straight into the provider and wire it in.

```bash
# overwrites a hand-written data source (retired from Datasources, moved to generated)
./bin/tfgen generate --spec $GEN_TEST/cost_budget.yaml
./bin/tfgen generate --spec $GEN_TEST/team.yaml
./bin/tfgen generate --spec $GEN_TEST/teams.yaml
./bin/tfgen generate --spec $GEN_TEST/incident_type.yaml
./bin/tfgen generate --spec $GEN_TEST/datastore.yaml
./bin/tfgen generate --spec $GEN_TEST/datastores.yaml
./bin/tfgen generate --spec $GEN_TEST/api_key.yaml
./bin/tfgen generate --spec $GEN_TEST/users.yaml

# purely additive (no hand-written counterpart)
./bin/tfgen generate --spec $GEN_TEST/api_keys.yaml
./bin/tfgen generate --spec $GEN_TEST/user.yaml
```

Generate every annotated slice in one sweep:

```bash
for f in $GEN_TEST/*.yaml; do ./bin/tfgen generate --spec "$f"; done
```

Verify and build:

```bash
gofmt -l datadog/fwprovider/data_source_datadog_*.go   # prints nothing if clean
make build
make test
```

---

## General commands (reuse later)

```bash
# Dry run: report what WOULD change, exit 3 if anything would, write nothing.
./bin/tfgen generate --spec $GEN_TEST/team.yaml --check

# Generate into a scratch dir instead of the live provider (inspect before wiring).
./bin/tfgen generate --spec $GEN_TEST/team.yaml --output-root .tfgen-out   # git-excluded

# Generate from a multi-artifact spec but only the named artifact(s).
./bin/tfgen generate --spec path/to/spec.yaml --include team,teams

# Write the run report to a file instead of stdout.
./bin/tfgen generate --spec $GEN_TEST/team.yaml --report /tmp/tfgen-report.json

# Point at the real V2 OpenAPI spec (the production default).
./bin/tfgen generate --spec .generator/V2/openapi.yaml --include <artifact_name>

# All flags.
./bin/tfgen generate --help
```

### Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--spec` | `.generator/V2/openapi.yaml` | OpenAPI spec to read |
| `--output-root` | `datadog/fwprovider` | Where data sources + registry files are written |
| `--include` | (all) | Comma-separated `artifact_name`s to generate |
| `--check` | `false` | Read-only; exit 3 if any file would change |
| `--report` | `-` (stdout) | Where to write the run report |
| `--tracking-field` | `x-datadog-tf-generator` | Extension name carrying the metadata |
| `--max-depth` | parser default | Hard limit on recursive `$ref` expansion |
| `--hooks-root` | `datadog/fwprovider/hooks` | Root for hook subpackages |

Exit codes: `0` success · `3` `--check` found changes · non-zero an artifact
failed (see the report).
