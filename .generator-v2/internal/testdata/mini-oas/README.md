# Mini OAS slices

Minimal, self-contained OpenAPI v2 specs — one per **singular** Datadog Terraform
data source (`mini-datadog_<name>.yaml`). Each file is a slice of the full Datadog
v2 OpenAPI spec containing only what that data source needs:

- the operation(s) the data source calls (the Go client method name is the
  `operationId`), e.g. `GetTeam` / `ListTeams`;
- the transitive `$ref` closure of those operations (schemas, parameters,
  responses);
- the security schemes used, with OAuth2 scopes trimmed to those referenced.

They are a corpus of **real Datadog response shapes** for exercising the generator
against production-shaped input, kept small enough to read and diff by hand. Every
file is a valid standalone OpenAPI 3.0.0 document (all `$ref`s resolve internally).

## Scope

V2 data sources only — data sources backed by the V1 API are excluded (no V1 spec
slice). The corpus includes unannotated slices and fixtures with tag registries.
A slice without a tag-level `x-datadog-tf-generator` registry declares no artifacts.

## Using a slice with the generator

For authors exercising the generator, add a family to the owning top-level Tag
Object. The family below declares both a by-id singular data source (`team`) and
a list data source (`teams`); retain only blocks whose operations are in the slice.

```yaml
tags:
- name: Teams
  x-datadog-tf-generator:
    team:
      datasource_singular:
        group:
          read: GetTeam
      datasource_list:
        group:
          read: ListTeams
```

See the [annotation reference](../../../docs/reference/annotation.md) for the
validated contract. Operation-level annotations are ignored.

```sh
tfgen generate --spec internal/testdata/mini-oas/mini-datadog_team.yaml \
  --output-root /tmp/out --report -
```

## Regenerating these slices

The scripts that produced this corpus live in `scripts/`:

- `_build_mini.py` — slices the singular V2 data sources (#4–#44). Run with no
  args for a recon table (writes nothing); pass `--build` to write the slices.
- `_build_mini_role.py` — slices the one-off `datadog_role` data source.
- `_validate.py` — checks every `mini-datadog_*.yaml` is a valid OpenAPI 3.0.0
  document with all `$ref`s resolving internally.
- `_annotate.py` — writes annotated copies of a representative sample (under
  `scripts/gen-test/`) for an end-to-end `tfgen generate` run. Each slice yields
  the variants its endpoints support: a `<name>.yaml` singular (by-id GET) and/or
  a `<plural>.yaml` list (collection GET, `datasource_list` block). Needs only the
  committed slices — not the full v2 spec.

The build scripts read the full v2 spec, which is **not** vendored here. Point at
it with the `DATADOG_OPENAPI_V2_SPEC` env var (defaults to
`~/local-dev/terraform/openapi.v2.yaml`); repo paths are derived from the script
location. The build scripts also read the provider's Go data source files under
`datadog/` to discover which operation(s) each data source calls.

```sh
cd scripts
DATADOG_OPENAPI_V2_SPEC=/path/to/openapi.v2.yaml python3 _build_mini.py --build
python3 _validate.py
```
