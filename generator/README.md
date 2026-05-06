# dd-tf-generator

A Go-based code generator that produces Terraform data source implementations from OpenAPI specifications.

## Build

```bash
cd generator
go build ./cmd/generator
```

## Usage

```bash
./generator generate --config config.yaml --output ../datadog/fwprovider/
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--config` | Yes | Path to YAML config file |
| `--output` | Yes | Output directory for generated Go files |
| `--dry-run` | No | Print generated code to stdout without writing files |

### Dry Run

Preview what would be generated without writing files:

```bash
./generator generate --config config.yaml --output /dev/null --dry-run
```

## Config Format

```yaml
specs:
  v2:
    path: ../.generator/V2/openapi.yaml   # Path to OpenAPI 3.0 spec (absolute or relative to config file)

datasources:
  team:
    spec: v2                               # Which spec to use (defaults to "v2")
    read:
      path: /api/v2/team/{team_id}         # API path from the spec
      method: get                          # HTTP method
```

### Filter-Fallback Lookup

Data sources can optionally include a `list` block to enable filter-based discovery. When configured, users can provide either the ID directly or use filter parameters to find the resource:

```yaml
datasources:
  team:
    spec: v2
    read:
      path: /api/v2/team/{team_id}
      method: get
    list:
      path: /api/v2/team
      method: get
```

With a `list` block:
- The path parameter (e.g., `team_id`) becomes Optional+Computed instead of Required
- Filter query parameters from the list endpoint (e.g., `filter[keyword]`) become Optional attributes
- A `ConfigValidators` method enforces that at least one of the ID or filter params is provided
- The generated Read method branches: ID-direct lookup when the ID is set, filter-based list lookup otherwise

### Fields

- **specs**: Map of spec names to their file paths. Paths can be absolute or relative to the config file directory.
- **datasources**: Map of data source names to their configuration.
  - **spec**: Reference to a spec defined in the `specs` section. Defaults to `"v2"`.
  - **read.path**: The API endpoint path as it appears in the OpenAPI spec.
  - **read.method**: The HTTP method (e.g., `get`).
  - **list.path** (optional): The list endpoint path for filter-fallback lookup.
  - **list.method** (optional): The HTTP method for the list endpoint.

## Output Files

For each data source (e.g., `team`), the generator produces two files:

| File | Overwritten on regeneration | Description |
|------|-----------------------------|-------------|
| `data_source_datadog_team_generated.go` | Yes | Auto-generated schema, model, Read(), and state mapping |
| `data_source_datadog_team_hooks.go` | No | Hook scaffold for custom logic; safe to edit |

The generated file includes a `DO NOT EDIT` header. The hooks file is only created if it doesn't already exist, preserving any customizations across regenerations.

## Testing

### Unit tests

```bash
cd generator
go test ./...
```

### Integration tests (go build gate)

Integration tests verify that generated code compiles within the provider using `go build`. They are guarded by a build tag and require the real Datadog V2 OpenAPI spec at `.generator/V2/openapi.yaml`:

```bash
cd generator
go test -tags=integration ./internal/codegen/
```

These tests generate data sources from the real spec, place them in the provider's `datadog/fwprovider/` directory, and run `go build` to catch type mismatches and invalid SDK accessor names that syntax-only parsing cannot detect.

## Supported Constructs

The generator handles:

- Flat JSON responses and JSON:API envelopes (auto-detected)
- Primitive types: string, integer (int32/int64), number (float/double), boolean
- Nested objects (generated as Terraform `SingleNestedBlock`)
- Arrays of objects (generated as Terraform `ListNestedBlock` with `NestedObject` wrapper)
- Arrays of primitives and arrays of objects
- Path parameters (mapped to Required attributes)
- Query parameters (mapped to Optional attributes)
- Response fields (mapped to Computed attributes)
- Nullable fields (use `GetXxxOk()` pattern with null fallback)
- Array-of-primitives fields (use `types.ListValueFrom`)
- `$ref`-based response types (derived from OpenAPI `$ref` names)
- Date-time fields are automatically excluded from generated code
- Filter-fallback lookup via optional `list` config block
- `allOf` flattening (property merge with last-wins semantics)
- `oneOf` / `anyOf` resolution (collect all variant properties as optional)
- `additionalProperties` mapped to `schema.MapAttribute` / `types.Map`

### Block types

The generator distinguishes between two Terraform block types:

- **`SingleNestedBlock`**: For singular nested objects. Attributes are placed directly on the block struct.
- **`ListNestedBlock`**: For arrays of objects. Attributes are placed inside a `NestedObject: schema.NestedBlockObject{}` wrapper (required by the Terraform Plugin Framework).

### SDK accessor casing

The Datadog Go SDK uses naive PascalCase for getter methods (e.g., `GetOrgId()` not `GetOrgID()`, `GetTeamUrl()` not `GetTeamURL()`). The generator produces SDK-compatible accessor names via `ToSDKPascalCase`, which capitalizes the first letter of each word segment without special acronym handling. Go struct field names in the model still use standard Go PascalCase with uppercase acronyms (e.g., `OrgID`, `TeamURL`).

### Composition types (allOf, oneOf, anyOf, additionalProperties)

The generator resolves OpenAPI composition constructs automatically:

- **`allOf`**: Properties from all sub-schemas are merged into a single flat schema. `$ref` targets are resolved before merging. Last definition wins on property name collisions. Required arrays are unioned.
- **`oneOf` / `anyOf`**: Properties from all variant schemas are collected as optional attributes. Compatible-type properties are merged; conflicting types fall back to string with a warning log.
- **`additionalProperties`**: When `true`, produces a `schema.MapAttribute` with `types.StringType` elements. When a typed schema, the element type is inferred (string, integer, number, boolean, or object).

Info-level logging records each composition resolution (e.g., `"Flattening allOf with 3 sub-schemas"`). For edge cases where the automatic resolution is insufficient, use the `ModifySchema` hook to override the generated schema per data source.
