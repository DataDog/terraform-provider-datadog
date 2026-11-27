# Phase 2a — Build the annotated slice with `slice_and_annotate.py`

The parameter set comes from Phase 1 (`references/input/collecting-inputs.md`); this file is
about *running* the script and understanding what it produces. Running tfgen on the result
is Phase 2b (`running-tfgen.md`).

## What the script does

`tfgen` generates artifacts declared in `tags[].x-datadog-tf-generator` family
registries. `slice_and_annotate.py` prepares one artifact from the full v2 spec:

1. **Slices** the full spec down to only the chosen operation(s) + their transitive `$ref`
   closure (schemas, params, responses, security schemes) — a valid standalone OpenAPI 3.0
   document.
2. **Declares** the requested artifact on the anchor operation's owning top-level
   tag, using the flags you pass. It removes other artifact declarations from the
   slice while preserving registry vendor metadata.

It writes the slice to a temp file and prints **only that path on stdout**, so you capture
it and feed it to tfgen. The script lives at
`.generator-v2/internal/testdata/mini-oas/scripts/slice_and_annotate.py`.

## How the annotation works

The contract is `.generator-v2/internal/contracts/family-registry.schema.json`;
its generated reference is `.generator-v2/docs/reference/annotation.md`.
The family key supplies the artifact name. Blocks are named `resource`,
`datasource_singular`, or `datasource_list`, and each block names its backing
operation IDs in `group`.

The script's `--artifact-kind` and `--cardinality` flags choose the block; they
are CLI inputs, not fields written into the registry. It sets an explicit
`artifact_name` to preserve the requested name, including plural names.
`tf_description` and `overwrites` are block-level fields.

`operationId` is the Go client method name (`GetTeam`, `ListTeams`).

## Running it

```bash
cd .generator-v2/internal/testdata/mini-oas/scripts

# stdout is ONLY the path; the human summary goes to stderr
SLICE=$(python3 slice_and_annotate.py \
  --spec "$SPEC" \
  --artifact-name team \
  --tf-description "Use this data source to retrieve information about an existing Datadog team." \
  --read GetTeam --search ListTeams)

echo "$SLICE"   # e.g. /var/folders/.../tfgen-slices/team.yaml
```

### Flag reference

| Flag | Required | Notes |
|---|---|---|
| `--spec` | no | Full v2 OAS (a local file). The skill supplies the curled spec from Phase 1; standalone, it defaults to `$DATADOG_OPENAPI_V2_SPEC`. |
| `--artifact-name` | **yes** | snake_case, no `datadog_` prefix. |
| `--artifact-kind` | no | `data_source` (default) or `resource`. |
| `--tf-description` | no | Doc string. |
| `--read` | one of read/search | by-id op (singular) **or** the list op (plural). |
| `--search` | one of read/search | list op for singular id-optional resolution. |
| `--create` / `--update` / `--delete` | no | Resource lifecycle ops. |
| `--cardinality` | no | `singular` (default) or `plural`. |
| `--overwrites` | no | Constructor to retire. |
| `--out` | no | Override output path. Default `$TMPDIR/tfgen-slices/<name>.yaml`. |

### Examples

```bash
# singular, read-only:
python3 slice_and_annotate.py --artifact-name incident_type \
  --tf-description "Use this data source to retrieve information about an existing Datadog incident type." \
  --read GetIncidentType

# singular, id-optional (by-id + list):
python3 slice_and_annotate.py --artifact-name team \
  --tf-description "Use this data source to retrieve information about an existing Datadog team." \
  --read GetTeam --search ListTeams

# plural (collection GET as the read):
python3 slice_and_annotate.py --artifact-name teams --cardinality plural \
  --tf-description "Use this data source to retrieve information about existing Datadog teams." \
  --read ListTeams

# overwrite a hand-written data source:
python3 slice_and_annotate.py --artifact-name team \
  --read GetTeam --search ListTeams \
  --overwrites NewDatadogTeamDataSource
```

## Errors the script raises (all exit non-zero)

| Message | Fix |
|---|---|
| `operationId(s) not found in spec: X` | Wrong id or wrong `--spec`; re-check in Phase 1. |
| `need at least one of --read or --search` | Provide the backing GET op. |
| `invalid --artifact-name 'X'` | Lowercase snake_case, ≤64. |
| `spec not found at ...` | Set `--spec` / `$DATADOG_OPENAPI_V2_SPEC`. |

## Notes for the PR (Phase 3)

- The slice is a **temp file** — an input to tfgen, not a committed artifact. Don't commit
  it. In the PR body, record the annotation params (operationIds, cardinality, overwrite)
  so the generation is reproducible; the committed artifacts are the generated `.go` files.
