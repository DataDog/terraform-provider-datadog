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
slice). These slices carry **no** `x-datadog-tf-generator` annotation, so the
generator skips them as-is; add the extension to opt an operation in.

## Using a slice with the generator

Annotate the read operation, then point `tfgen generate` at the file:

```yaml
# under paths./api/v2/team/{team_id}.get:
x-datadog-tf-generator:
  artifact_kind: data_source
  artifact_name: team
  group:
    read: GetTeam
```

```sh
tfgen generate --spec internal/testdata/mini-oas/mini-datadog_team.yaml \
  --output-root /tmp/out --report -
```
