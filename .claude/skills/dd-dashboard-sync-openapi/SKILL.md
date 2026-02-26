---
name: dd-dashboard-sync-openapi
description: >
  Syncs the datadog_dashboard_v2 Terraform resource with the Datadog OpenAPI spec.
  Identifies fields and widget types present in the OpenAPI spec but missing from
  the FieldSpec-based implementation, generates the Go additions, writes and records
  new acceptance tests, then opens a PR. Use this skill whenever the user mentions
  syncing dashboard fields with the API spec, adding missing widget types, checking
  for OpenAPI gaps, updating dashboard_v2 FieldSpecs, or diffing the dashboard resource
  against the spec — even if they don't explicitly say "OpenAPI sync". Also use when the
  user asks to add a specific widget type or field group to dashboard_v2. Requires
  DD_TEST_CLIENT_API_KEY and DD_TEST_CLIENT_APP_KEY to be set (for RECORD=true cassette
  recording).
tools: Bash, Read, Write, Edit, Glob, Grep
model: sonnet
---

# Dashboard OpenAPI Sync Skill

You are syncing the `datadog_dashboard_v2` Terraform resource with the Datadog OpenAPI spec.
The resource uses a FieldSpec-based bidirectional mapping system described in AGENTS.md.
Read that file first for the full conventions before proceeding.

## Inputs

The user may optionally specify:
- A specific widget type (e.g. "timeseries", "toplist") or schema name (e.g. "LogQueryDefinition") to focus on
- "all" or no input: show all gaps and proceed interactively

## Step 1 — Orient

Read these files before doing anything else:
- `AGENTS.md` (conventions, FieldSpec system, naming rules)
- `datadog/dashboardmapping/widgets.go` (WidgetSpec declarations and `allWidgetSpecs` registry)
- `datadog/dashboardmapping/field_groups.go` (shared reusable FieldSpec groups)
- `datadog/dashboardmapping/field_groups_dashboard.go` (dashboard top-level field groups — NOT shared with widget specs)
- The OpenAPI spec (source of truth): `https://github.com/DataDog/datadog-api-spec/blob/master/spec/v1/dashboard.yaml`

## Step 2 — Diff: Identify Gaps

### 2a. Widget types

In `dashboard.yaml`, find the `WidgetDefinition` schema's `oneOf` list — every `$ref` there
is a widget type the API supports. For each widget type:
1. Derive the `JSONType` string (e.g. `TimeseriesWidgetDefinition` → `"timeseries"`)
2. Check whether a `WidgetSpec` with that `JSONType` exists in `allWidgetSpecs` in `widgets.go`
3. Record missing widget types

### 2b. Fields on existing FieldSpec groups

For each reusable FieldSpec group in `field_groups.go` (named after its OpenAPI counterpart, e.g.
`logQueryDefinitionFields` → `LogQueryDefinition`):
1. Find the corresponding OpenAPI schema in `dashboard.yaml`
2. List its properties
3. Compare against the FieldSpec group's `HCLKey` entries (accounting for renames — check
   comments and `JSONKey` overrides)
4. Record missing properties

### 2c. Fields on existing widget specs

For each `WidgetSpec` in `allWidgetSpecs`, find its OpenAPI schema (e.g. `TimeseriesWidgetDefinition`)
and check for properties not yet mapped. When focusing on a specific widget, also check the shared
field groups it references — a missing field on a shared group (e.g. `WidgetMarker.time`) affects
every widget that uses that group.

### 2d. Engine post-processing

For formula-capable widgets, check whether `buildFormulaRequest` / `flattenFormulaRequest`
in `engine.go` handle all the fields needed. These functions are driven by `FormulaRequestConfig`
— check the per-widget config (`timeseriesFormulaRequestConfig`, `scalarFormulaRequestConfig`,
etc.) to see which style fields and extra request fields are declared. If a newly-added field
belongs at the request level (not inside a formula or query), the relevant config's `ExtraFields`
or `StyleFields` may need updating.

### 2e. Present findings

Show the user a structured gap report:
```
MISSING WIDGET TYPES:
  - bar_chart (BarChartWidgetDefinition)
  - funnel    (FunnelWidgetDefinition)

MISSING FIELDS ON EXISTING GROUPS:
  - LogQueryDefinition.multi_compute  (not yet mapped in logQueryDefinitionFields)
  - WidgetAxis.label                  (present in spec, confirm OmitEmpty behavior)

MISSING FIELDS ON EXISTING WIDGETS:
  - timeseries: semantic_mode on FormulaAndFunctionMetricQueryDefinition
```

Ask the user which gaps to address in this session (all, a subset, or a specific widget).

## Step 3 — Design Additions

For each gap the user wants to address:

### New field on an existing FieldSpec group

1. Read the OpenAPI property definition (type, required/optional, description)
2. Determine `OmitEmpty`:
   - If the property is in the schema's `required` array → `OmitEmpty: false`
   - If optional → `OmitEmpty: true` (default; flag for cassette verification)
3. Determine `Type` from OpenAPI type:
   - `string` → `TypeString`
   - `boolean` → `TypeBool`
   - `integer` → `TypeInt`
   - `array` of strings/ints → `TypeStringList` / `TypeIntList`
   - `array` of objects → `TypeBlockList` (with Children)
   - `object` (or `$ref` to object) → `TypeBlock` (with Children)
   - `oneOf` with a JSON discriminator field → `TypeOneOf` (see below)
4. Determine `JSONKey`:
   - If the HCL name differs from the OpenAPI property name, set `JSONKey`
   - Apply singular/plural rule: HCL uses singular block names, JSON uses the OpenAPI (plural) key
5. Set `Description` from the OpenAPI property's `description` field
6. Set `ValidValues` from the OpenAPI enum values if the property is an enum type

Ask the user to confirm the `OmitEmpty` decision for any field that is optional but
whose cassette behavior is unknown.

### OneOf fields: use TypeOneOf

When an OpenAPI property is a `oneOf` (its variants share a JSON object location but differ in
shape, usually distinguished by a `type` discriminator field), use `TypeOneOf` instead of
`TypeBlock`. Do NOT write `SchemaOnly + post-process hook` code for discriminated unions.

**TypeOneOf structure:**
```go
{HCLKey: "field_name", Type: TypeOneOf, OmitEmpty: true,
    Description: "...",
    Discriminator: &OneOfDiscriminator{JSONKey: "type"},  // discriminator field in JSON
    Children: []FieldSpec{
        {HCLKey: "variant_a", Type: TypeBlock, OmitEmpty: true,
            Discriminator: &OneOfDiscriminator{Value: "type_value_a"},  // injected on build
            Children: variantAFields},
        {HCLKey: "variant_b", Type: TypeBlock, OmitEmpty: true,
            Discriminator: &OneOfDiscriminator{Value: "type_value_b"},
            Children: variantBFields},
    },
}
```

**How it works:**
- Build (HCL→JSON): finds the populated child block, builds its JSON, injects `{"type": "type_value_a"}` automatically
- Flatten (JSON→HCL): reads `json["type"]`, matches against `Discriminator.Value`, populates only the matching child
- If a variant maps to multiple discriminator values (e.g. "table" or "none"), use `Discriminator.Values: []string{"table", "none"}` — no build injection
- If the legacy JSON has no discriminator field, mark that variant `DefaultVariant: true`

**Example:** `NumberFormatUnit` (canonical vs custom) in `widgetNumberFormatFields` — the `unit` field is TypeOneOf.

**When NOT to use TypeOneOf:** if the oneOf variants each require custom build/flatten logic beyond field mapping (e.g. `FormulaAndFunctionQueryDefinition`), keep the existing custom engine functions.

### New widget type

1. Read the widget's OpenAPI schema (e.g. `BarChartWidgetDefinition`)
2. Identify which properties are covered by existing reusable FieldSpec groups in `field_groups.go`
   (check for `$ref` to known schemas like `WidgetCustomLink`, `WidgetTime`, `WidgetAxis`, etc.)
3. Design new per-widget FieldSpec entries for properties not covered by shared groups
4. Write the `WidgetSpec` struct and register it in `allWidgetSpecs` in `widgets.go`

Note: do NOT manually modify `fwprovider/resource_datadog_dashboard_v2.go`. Adding the `WidgetSpec`
to `allWidgetSpecs` is sufficient — `AllWidgetFWBlocks` and `WidgetSpecToFWBlock` in `schema_gen.go`
generate the Terraform framework schema and state converters automatically. `AllWidgetAttrTypes` also
auto-updates to include the new widget in state conversion.

The FieldSpec engine serializes directly to/from `map[string]interface{}` JSON — it does NOT use
the generated Go API client types (e.g. `datadogV1.TimeseriesWidgetDefinition`). Any widget or field
that exists in the OpenAPI spec can be implemented regardless of whether the Go API client has
generated types for it.

Pause and show the proposed design to the user before writing code.

## Step 4 — Implement

Write the Go code additions:
1. Add new `FieldSpec` group variables to `field_groups.go` if a new reusable group is needed
2. Add or update `WidgetSpec` entries in `widgets.go`

**Schema is auto-generated from FieldSpec — no manual schema entries are needed.**
`FieldSpecToFWAttribute` and `FieldSpecsToFWSchema` in `schema_gen.go` convert each `FieldSpec`
to a framework `schema.Attribute` or `schema.Block` automatically, using `Description`,
`ValidValues`, `Required`, `Default`, etc.

3. If a new widget type requires formula/query request support, add a `FormulaRequestConfig` entry
   in `engine.go` and register it in `formulaRequestConfigForWidget`. Do NOT write new
   `buildFormulaQueryRequestJSON`-style functions — the unified `buildFormulaRequest` /
   `flattenFormulaRequest` handles all formula-capable widgets through the config.
   For non-formula post-processing (injected constants, recursive widget dispatch), update
   `buildWidgetPostProcess` and `flattenWidgetPostProcess` in `engine.go`.

4. Run `go build ./...` and `go vet ./datadog/dashboardmapping/...` after writing.

## Step 5 — Write Acceptance Test

For each new widget type or set of new fields, create a new test file with its own cassette.
Do not reuse v1 cassettes — the v1 `datadog_dashboard` resource is not being maintained.

Create `datadog/tests/resource_datadog_dashboard_v2_{widget}_test.go`:

```go
package test

import (
    "testing"
)

const datadogDashboard{Widget}Config = `
resource "datadog_dashboard_v2" "{widget}_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        {widget}_definition {
            // ... all fields ...
        }
    }
}
`

var datadogDashboard{Widget}Asserts = []string{
    "title = {{uniq}}",
    "widget.0.{widget}_definition.0.some_field = expected_value",
    // ... cover every new field ...
}

func TestAccDatadogDashboardV2{Widget}(t *testing.T) {
    config, name := datadogDashboard{Widget}Config, "datadog_dashboard_v2.{widget}_dashboard"
    testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2{Widget}", config, name, datadogDashboard{Widget}Asserts)
}

func TestAccDatadogDashboardV2{Widget}_import(t *testing.T) {
    config, name := datadogDashboard{Widget}Config, "datadog_dashboard_v2.{widget}_dashboard"
    testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2{Widget}_import", config, name)
}
```

Key points:
- Config is a **`const` string**, not a function
- No extra imports beyond `"testing"`
- `{{uniq}}` is the placeholder for the unique dashboard name
- Uses `testAccDatadogDashboardV2WidgetUtil` (framework mux provider)
- The cassette name argument (`"TestAccDatadogDashboardV2{Widget}"`) must match the test name
  exactly so `testClockWithName` picks up the right `.freeze` file

### Register in provider_test.go

Add the new test file to `datadog/tests/provider_test.go` in the `testFiles2EndpointTags` map:
```go
"tests/resource_datadog_dashboard_v2_{widget}_test": "dashboards",
```
Without this entry the test will fail with:
`Endpoint tag for test file ... not found in datadog/provider_test.go`

## Step 6 — Record Cassettes (RECORD=true)

**Prerequisites — two environment issues must be fixed before running tests:**

1. **Export API keys** — `make testacc` only inherits env vars that are exported:
   ```bash
   export DD_TEST_CLIENT_API_KEY DD_TEST_CLIENT_APP_KEY
   ```

2. **Unset `OTEL_TRACES_EXPORTER`** — if set to `otlp`, `terraform version` exits with code 1
   and every acceptance test fails immediately. Clear it inline on the make invocation.

**Record cassettes:**
```bash
export DD_TEST_CLIENT_API_KEY DD_TEST_CLIENT_APP_KEY
OTEL_TRACES_EXPORTER= RECORD=true TESTARGS="-run TestAccDatadogDashboardV2{Widget}$" make testacc
```

**Verify cassette replay passes:**
```bash
OTEL_TRACES_EXPORTER= RECORD=false TESTARGS="-run TestAccDatadogDashboardV2{Widget}" make testacc
```

If `RECORD=false` replay fails after a successful `RECORD=true` run, check:
1. `OmitEmpty` — a field being included/excluded incorrectly
2. `JSONKey` — missing or wrong singular/plural conversion
3. `Default` — fields with `Default` are always emitted even when not set in HCL
4. `SchemaOnly` — fields that must not be serialized to JSON need `SchemaOnly: true`
5. **List ordering** — `UseSet: true` fields use `ListAttribute` (not `SetAttribute`) to
   preserve HCL insertion order; the cassette body will reflect whatever order the user
   wrote in the config, so be consistent

## Step 7 — Quality Gates

```bash
# Build and vet
go build ./...
go vet ./datadog/dashboardmapping/...

# Docs: make docs requires terraform and may fail due to OTEL env issues.
# make check-docs does NOT require terraform and reliably verifies docs are in sync.
OTEL_TRACES_EXPORTER= make check-docs
```

All must pass before creating the PR.

## Step 8 — Create PR

Branch name: `{github-username}/dashboard-{widget-or-schema}`

PR title: `[datadog_dashboard_v2] Add {description} from OpenAPI sync`

PR body should include:
- Which OpenAPI schema version / commit was diffed against
- List of widgets/fields added
- Note that cassettes were recorded with `RECORD=true` and verified with `RECORD=false`
- Label: `improvement`

## Constraints

- Never re-record cassettes for **existing** tests — only record for newly added tests
- Existing test assertions must not change
- All new FieldSpec entries must have `Description` set (drives `make docs`)
- `OmitEmpty` for new optional fields defaults to `true`; flag any that may need cassette
  verification in a PR comment
- The OpenAPI spec URL is:
  `https://github.com/DataDog/datadog-api-spec/blob/master/spec/v1/dashboard.yaml`
- FieldSpec declarations live in:
  - `datadog/dashboardmapping/field_groups.go` — shared groups
  - `datadog/dashboardmapping/field_groups_dashboard.go` — dashboard top-level groups
  - `datadog/dashboardmapping/widgets.go` — widget specs and registry
- Framework schema generation is fully automatic via `schema_gen.go` — never modify
  `fwprovider/resource_datadog_dashboard_v2.go` to add widget-level schema
