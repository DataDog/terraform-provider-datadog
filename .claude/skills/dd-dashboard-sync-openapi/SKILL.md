---
name: dd-dashboard-sync-openapi
description: >
  Syncs the datadog_dashboard Terraform resource with the Datadog OpenAPI spec.
  Identifies fields and widget types present in the OpenAPI spec but missing from
  the FieldSpec-based implementation, generates the Go additions, writes and records
  new acceptance tests, then opens a PR. Requires DD_TEST_CLIENT_API_KEY and
  DD_TEST_CLIENT_APP_KEY to be set (for RECORD=true cassette recording).
tools: Bash, Read, Write, Edit, Glob, Grep
model: sonnet
---

# Dashboard OpenAPI Sync Skill

You are syncing the `datadog_dashboard` Terraform resource with the Datadog OpenAPI spec.
The resource uses a FieldSpec-based bidirectional mapping system described in AGENTS.md.
Read that file first for the full conventions before proceeding.

## Inputs

The user may optionally specify:
- A specific widget type (e.g. "timeseries", "toplist") or schema name (e.g. "LogQueryDefinition") to focus on
- "all" or no input: show all gaps and proceed interactively

## Step 1 — Orient

Read these files before doing anything else:
- `AGENTS.md` (conventions, FieldSpec system, naming rules)
- `datadog/resource_datadog_dashboard_new.go` (current FieldSpec definitions and WidgetSpec registry)
- `/Users/andy.yacomink/go/src/github.com/DataDog/datadog-api-spec/spec/v1/dashboard.yaml` (OpenAPI source of truth)

## Step 2 — Diff: Identify Gaps

### 2a. Widget types

In `dashboard.yaml`, find the `WidgetDefinition` schema's `oneOf` list — every `$ref` there
is a widget type the API supports. For each widget type:
1. Derive the `JSONType` string (e.g. `TimeseriesWidgetDefinition` → `"timeseries"`)
2. Check whether a `WidgetSpec` with that `JSONType` exists in `allWidgetSpecs` in the resource file
3. Record missing widget types

### 2b. Fields on existing FieldSpec groups

For each reusable FieldSpec group in the resource (named after its OpenAPI counterpart, e.g.
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
5. Add a comment citing the OpenAPI property name if the names differ

Ask the user to confirm the `OmitEmpty` decision for any field that is optional but
whose cassette behavior is unknown.

### New widget type

1. Read the widget's OpenAPI schema (e.g. `BarChartWidgetDefinition`)
2. Identify which properties are covered by existing reusable FieldSpec groups
   (check for `$ref` to known schemas like `WidgetCustomLink`, `WidgetTime`, `WidgetAxis`, etc.)
3. Design new per-widget FieldSpec entries for properties not covered by shared groups
4. Write the `WidgetSpec` struct, registering it in `allWidgetSpecs`
5. Add the corresponding HCL schema block to `resourceDatadogDashboard()`

Pause and show the proposed design to the user before writing code.

## Step 4 — Implement

Write the Go code additions:
1. Add new `FieldSpec` group variables (if a new reusable group is needed)
2. Add or update `WidgetSpec` entries
3. Add HCL schema fields (`schema.Schema`) for any new fields or widget types
4. All schema entries MUST have `Description` fields (required by `make docs`)
5. Run `make fmtcheck` and `make test` after writing

## Step 5 — Write Acceptance Test

For each new widget type or significantly new set of fields:

1. Find the appropriate test file: `datadog/tests/resource_datadog_dashboard_{widget}_test.go`
   (create it if it doesn't exist for a new widget type)
2. Write a test function following the existing pattern:
   - HCL config using `{{uniq}}` for the dashboard name
   - Assertions array covering every new field
   - Test function calling `testAccDatadogDashboardWidgetUtil`
3. For new widget test files, add the import block following existing test files

Refer to existing test files (e.g. `resource_datadog_dashboard_timeseries_test.go`) as templates.

## Step 6 — Record Cassettes (RECORD=true)

```bash
# Verify API keys are set
if [ -z "$DD_TEST_CLIENT_API_KEY" ] || [ -z "$DD_TEST_CLIENT_APP_KEY" ]; then
  echo "ERROR: DD_TEST_CLIENT_API_KEY and DD_TEST_CLIENT_APP_KEY must be set for cassette recording"
  exit 1
fi

# Record cassettes for new/changed tests
RECORD=true \
  DD_TEST_CLIENT_API_KEY=$DD_TEST_CLIENT_API_KEY \
  DD_TEST_CLIENT_APP_KEY=$DD_TEST_CLIENT_APP_KEY \
  TESTARGS="-run TestAccDatadogDashboard{WidgetName}" \
  make testacc
```

After recording, run with `RECORD=false` to confirm cassette replay passes:
```bash
RECORD=false TESTARGS="-run TestAccDatadogDashboard{WidgetName}" make testacc
```

If `RECORD=false` replay fails after a successful `RECORD=true` run, the likely cause is
non-deterministic JSON serialization. Check the cassette body against the generated request
body (see AGENTS.md — Debugging Cassette Mismatches).

## Step 7 — Quality Gates

```bash
make fmtcheck
make test
make vet
make errcheck
make docs && make check-docs
```

All must pass before creating the PR.

## Step 8 — Create PR

Branch name: `yacomink/YYYYMMDD-dashboard-openapi-sync-{widget-or-schema}`

PR title: `[datadog_dashboard] Add {description} from OpenAPI sync`

PR body should include:
- Which OpenAPI schema version / commit was diffed against
- List of widgets/fields added
- Note that cassettes were recorded with `RECORD=true` and verified with `RECORD=false`
- Label: `improvement`

## Constraints

- Never re-record cassettes for **existing** tests — only record for newly added tests
- Existing test assertions must not change
- All new schema fields must have `Description` set
- `OmitEmpty` for new optional fields defaults to `true`; flag any that may need cassette
  verification in a PR comment
- The OpenAPI spec path is always:
  `/Users/andy.yacomink/go/src/github.com/DataDog/datadog-api-spec/spec/v1/dashboard.yaml`
- The resource implementation is always: `datadog/resource_datadog_dashboard_new.go`
  (adjust if the file was renamed)
