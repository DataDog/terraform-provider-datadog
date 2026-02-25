---
name: dd-dashboard-sync-openapi
description: >
  Syncs the datadog_dashboard_v2 Terraform resource with the Datadog OpenAPI spec.
  Identifies fields and widget types present in the OpenAPI spec but missing from
  the FieldSpec-based implementation, generates the Go additions, writes and records
  new acceptance tests, then opens a PR. Requires DD_TEST_CLIENT_API_KEY and
  DD_TEST_CLIENT_APP_KEY to be set (for RECORD=true cassette recording).
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
- `https://github.com/DataDog/datadog-api-spec/blob/master/spec/v1/dashboard.yaml` (OpenAPI source of truth)

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
and check for properties not yet mapped.

### 2d. Present findings

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
4. Determine `JSONKey`:
   - If the HCL name differs from the OpenAPI property name, set `JSONKey`
   - Apply singular/plural rule: HCL uses singular block names, JSON uses the OpenAPI (plural) key
5. Set `Description` from the OpenAPI property's `description` field
6. Set `ValidValues` from the OpenAPI enum values if the property is an enum type

Ask the user to confirm the `OmitEmpty` decision for any field that is optional but
whose cassette behavior is unknown.

### New widget type

1. Read the widget's OpenAPI schema (e.g. `BarChartWidgetDefinition`)
2. Identify which properties are covered by existing reusable FieldSpec groups in `field_groups.go`
   (check for `$ref` to known schemas like `WidgetCustomLink`, `WidgetTime`, `WidgetAxis`, etc.)
3. Design new per-widget FieldSpec entries for properties not covered by shared groups
4. Write the `WidgetSpec` struct and register it in `allWidgetSpecs` in `widgets.go`

Note: do NOT manually add entries to `resourceDatadogDashboard()`. Adding the `WidgetSpec`
to `allWidgetSpecs` is sufficient — `AllWidgetSchemasMap` and `WidgetSpecToSchemaBlock`
generate the Terraform schema automatically.

Pause and show the proposed design to the user before writing code.

## Step 4 — Implement

Write the Go code additions:
1. Add new `FieldSpec` group variables to `field_groups.go` if a new reusable group is needed
2. Add or update `WidgetSpec` entries in `widgets.go`

**Schema is auto-generated from FieldSpec — no manual `schema.Schema` entries are needed.**
`FieldSpecToSchemaElem` in `schema_gen.go` converts each `FieldSpec` to a `*schema.Schema`
automatically, using `Description`, `ValidValues`, `Required`, `Default`, etc.

3. Run `make fmtcheck` and `make test` after writing

## Step 5 — Write Acceptance Test

For each new widget type or significantly new set of fields:

1. Find the appropriate test file: `datadog/tests/resource_datadog_dashboard_v2_{widget}_test.go`
   (create it if it doesn't exist for a new widget type)

2. Write the test file. Pattern (see `resource_datadog_dashboard_slo_list_test.go` as a minimal template):
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

   func TestAccDatadogDashboard{Widget}(t *testing.T) {
       testAccDatadogDashboardWidgetUtil(t, datadogDashboard{Widget}Config, "datadog_dashboard_v2.{widget}_dashboard", datadogDashboard{Widget}Asserts)
   }

   func TestAccDatadogDashboard{Widget}_import(t *testing.T) {
       testAccDatadogDashboardWidgetUtilImport(t, datadogDashboard{Widget}Config, "datadog_dashboard_v2.{widget}_dashboard")
   }
   ```
   Key points:
   - Config is a **`const` string**, not a function
   - No extra imports beyond `"testing"`
   - `{{uniq}}` is the placeholder for the unique dashboard name

3. **Register the new test file** in `datadog/tests/provider_test.go` in the `testFiles2EndpointTags` map.
   Find the alphabetical position among other `dashboard_*` entries and add:
   ```go
   "tests/resource_datadog_dashboard_v2_{widget}_test": "dashboards",
   ```
   Without this entry the test will immediately fail with:
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
OTEL_TRACES_EXPORTER= RECORD=true TESTARGS="-run TestAccDatadogDashboard{Widget}$" make testacc
```

**Verify cassette replay passes:**
```bash
OTEL_TRACES_EXPORTER= RECORD=false TESTARGS="-run TestAccDatadogDashboard{Widget}" make testacc
```

If `RECORD=false` replay fails after a successful `RECORD=true` run, check:
1. `OmitEmpty` — a field being included/excluded incorrectly
2. `JSONKey` — missing or wrong singular/plural conversion
3. `Default` — fields with `Default` are always emitted by Terraform even when not set in HCL
4. `SchemaOnly` — fields that must not be serialized to JSON need `SchemaOnly: true`

## Step 7 — Quality Gates

```bash
# Build and vet (make fmtcheck fails on pre-existing example formatting issues — skip it)
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
- The OpenAPI spec path is always:
  `https://github.com/DataDog/datadog-api-spec/blob/master/spec/v1/dashboard.yaml`
- FieldSpec declarations live in:
  - `datadog/dashboardmapping/field_groups.go` — shared groups
  - `datadog/dashboardmapping/field_groups_dashboard.go` — dashboard top-level groups
  - `datadog/dashboardmapping/widgets.go` — widget specs and registry
