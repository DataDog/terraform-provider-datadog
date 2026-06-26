---
name: tfgen-vacuum
description: Generate a Terraform provider data source (.go) from an OpenAPI spec snippet using the generator-v2 emit templates, as a STRICT vacuum experiment. Invoke when the user provides an OpenAPI YAML snippet (with an x-datadog-tf-generator block) and asks to generate a data source. You render the templates from the spec ALONE — you must never read any existing provider implementation, generator Go code, SDK source, or prior generated output. Everything is derived by reasoning about the spec + the templates.
---

# tfgen-vacuum — generate a data source in a vacuum

You are acting as the generator-v2 emit stage **by hand**. Given an OpenAPI spec
snippet and the emit templates, you produce a working Terraform-plugin-framework
data source `.go` file.

This is a controlled experiment. Its entire value depends on you generating the
output from **only the spec + the templates**, reasoning everything else out
yourself. If you peek at how the answer was done elsewhere, the experiment is
worthless.

## 🛑 THE VACUUM RULE — read this first, it is the whole point

You **MUST NOT** look at any existing implementation, answer key, or prior result
to inform the output. This includes — do not Read, Grep, Glob, `cat`, open in an
editor, or fan out an `Agent`/`Explore`/`general-purpose` search over any of:

- **Any provider implementation** — anything under `datadog/` (especially
  `datadog/fwprovider/data_source_*.go`, `datadog/fwprovider/resource_*.go`, and
  any hand-written or already-generated data source/resource).
- **The generator's own Go code** — anything under `.generator-v2/**/*.go`
  (parser, model, emit, binder, etc.). That code computes the very template
  variables you are being tested on. It is the answer key. Off-limits.
- **The Datadog Go SDK source** — `datadog-api-client-go` / any `vendor/` copy.
  You derive SDK names (packages, API structs, methods, getters) by reasoning
  about OpenAPI → Go-client naming conventions (see Conventions below), not by
  looking them up.
- **Prior generated output** — the *contents* of any file under `.tfgen-out/`.
  You may LIST filenames there (to compute the iteration number) but never read
  their contents, even your own previous iterations.

You may ONLY read:

1. The **OpenAPI YAML snippet** the user provides (inline, or a file path they
   give you — that file is the spec, so Read it).
2. The **template files** in
   `/Users/jason.tenczar/projects/terraform-provider-datadog/.generator-v2/internal/emit/templates/`
   (`data_source_common.go.tmpl`, `data_source_singular.go.tmpl`,
   `data_source_plural.go.tmpl`).
3. Filenames only (not contents) under `.tfgen-out/claude/<type>/`, to count
   iterations.

If you are ever unsure whether a read is allowed: if it isn't the spec, a
template, or a directory listing for the iteration count — **don't**. Reason it
out instead. In your final reasoning trace you will attest to exactly what you
read.

## Inputs

The user provides an OpenAPI spec snippet (YAML) containing one or more `paths`
operations and an `x-datadog-tf-generator` block on the artifact's primary
operation. If no snippet is present in the request, ask for it and stop.

## Procedure

### Step 1 — Read the spec and the templates

Read the YAML snippet. Read all three template files (you always need
`data_source_common.go.tmpl` — it defines the shared partials `modelStructs`,
`schemaAttribute`, `nestedBody`, `schemaBlock`, `boilerplate` — plus whichever of
`singular`/`plural` you select). The template comment headers describe when each
shape applies; rely on them.

If `x-datadog-tf-generator.artifact_kind` is not `data_source`, stop: there is no
template for it yet. Report that and do nothing else.

### Step 2 — Classify singular vs plural

Determine the data source SHAPE from the spec alone:

- **singular** — resolves **exactly one** record into flat top-level state.
  Signals: `artifact_name` is a singular noun; `x-datadog-tf-generator.group` has
  a by-id `read` and/or a `search` that narrows to one; the artifact is "get the
  X with this id / matching these filters".
- **plural** — projects a **collection** into a repeated nested block (a list of
  items in state). Signals: `artifact_name` is a plural noun; the primary
  operation is a List whose items are all kept; state holds many records.

Note that a singular "both" data source also references a List operation (as its
`search`) — an array response alone does **not** make it plural. The question is
whether final state holds **one** record (singular) or **many** (plural).

For **singular**, resolve the sub-mode from `group`:
- by-id `read` only → **read-only**: `Searchable=false`, `ByID` n/a. Imports
  `fmt`; Read calls the by-id method and 404s with a not-found error.
- `search`/list only (no by-id read) → **search-only**: `Searchable=true`,
  `ByID=false`. `id` attribute is `utils.ResourceIDAttribute()`.
- both `read` (by-id) and `search` (list) → **both**: `Searchable=true`,
  `ByID=true`. `id` attribute is `Optional`+`Computed`.

### Step 3 — Build the template data model by reasoning

Walk the spec and compute every variable the chosen template + common partials
reference. Use the Conventions section below. Key pieces:

- Identify the operation(s) via `x-datadog-tf-generator.group` (or the single
  operation for a plain list). Method names are the `operationId`s.
- Resolve `$ref`s to find the response envelope and the item schema. Apply
  **envelope flattening** (Conventions).
- For each leaf field, map OpenAPI type → framework attribute type, model field
  type, SDK getter, and the state-assignment RHS.
- Nested objects/arrays become blocks and additional structs in `.Models` (use
  the common template's `schemaBlock`/`nestedBody`).

### Step 4 — Render to Go

Produce the full file exactly as the templates would, substituting your computed
values. The real generator runs the output through `go/format`, so emit
**gofmt-clean** Go by hand: tabs for indentation, aligned struct fields, no stray
blank lines. The code must be reasoned to **compile and work** — correct package,
imports actually used (and only those the template's import block allows; e.g.
the singular import block never includes `time`), consistent identifiers between
the model struct, schema, Read, and `updateState`.

Do **not** run `go`, `gofmt`, `make`, or any build/format command — keep it pure,
and the repo's convention is to never run raw go commands anyway. Get the
formatting right by hand.

### Step 5 — Determine output path + iteration number

Output goes in a directory named for the **type** (shape) you generated:

```
/Users/jason.tenczar/projects/terraform-provider-datadog/.tfgen-out/claude/<type>/<name>_<N>.go
```

- `<type>` is `singular` or `plural`.
- `<name>` is `x-datadog-tf-generator.artifact_name`.
- `<N>` is the iteration number, **scoped to that type directory**: list filenames
  matching `<name>_*.go` in `.tfgen-out/claude/<type>/`, take the highest existing
  index, add 1 (start at 1 if none). Iterations are per-type — a plural run does
  not continue a singular run's count. (e.g. if `singular/datastore_1.go` exists
  and you now generate the plural datastore, it is `plural/datastore_1.go`.)

List filenames only — never read their contents. The Write tool creates the
`<type>` directory if missing.

### Step 6 — Write the file and emit a reasoning trace

Write the `.go` file, then in your reply to the user print a short **reasoning
trace** so the experiment is auditable:

- The classification (singular/plural and sub-mode) and the signals that decided it.
- A decision table of the resolved template variables (SDKPackage, APIStruct,
  APIAccessor, GoName, TypeName, method names, Searchable/ByID or
  Paginated/Filters, item/state types).
- Any non-obvious type decisions (date-time `.String()`, enum `string()` casts,
  envelope flattening choices, dropped `type` discriminator, blocks).
- An explicit **attestation**: "Did not read: provider implementations,
  generator Go code, SDK source, or prior output. Read only: the spec + the
  templates (+ a filename listing for the iteration count)."
- Flag the assumptions a reviewer should double-check (since you did not verify
  against the SDK), e.g. the exact getter/method/accessor names.

## Conventions (derive by reasoning — never look these up in source)

**SDK identifiers** (openapi-generator Go client naming):
- `SDKPackage`: API version from the paths — `/api/v2/…` → `datadogV2`,
  `/api/v1/…` → `datadogV1`.
- `APIStruct`: operation `tags[0]`, PascalCased with spaces/punctuation removed,
  + `Api`. `Actions Datastores` → `ActionsDatastoresApi`.
- `APIAccessor`: `Get` + that PascalCased tag + `Api` + version. →
  `GetActionsDatastoresApiV2`. (Called as `providerData.DatadogApiInstances.<APIAccessor>()`.)
- Operation methods = the `operationId` (`GetDatastore`, `ListDatastores`).
  Paginated list helper is `<Method>WithPagination`.
- Field accessors: `Get<Field>Ok() (*T, bool)`, `Get<Field>()`, `Has<Field>()`,
  where `<Field>` title-cases each underscore-separated word and concatenates:
  `created_at`→`CreatedAt`, `org_id`→`OrgId`, `creator_user_uuid`→`CreatorUserUuid`,
  `id`→`Id`.

**Naming**:
- `GoName`: `artifact_name` as a lowercase (camelCase if multi-word) identifier —
  used for the *unexported* `<GoName>DataSource` struct and
  `<GoName>DataSourceModel`; the constructor is `New<Title GoName>DataSource`.
- `TypeName`: `artifact_name` (used in `response.TypeName` and the `id`
  description). `Description`: `x-datadog-tf-generator.tf_description`.
- Model struct field: PascalCase of the json/tf name; use `ID` for `id` in the
  model (but the SDK getter is `GetIdOk`).

**Type mapping** (OpenAPI → framework attr / model type / SDK getter / RHS):
- `string` → `schema.StringAttribute` / `types.String` / `*string` /
  `types.StringValue(*v)`.
- `integer` `int64` → `schema.Int64Attribute` / `types.Int64` / `*int64` /
  `types.Int64Value(*v)`. (`int32` → `Int32` analogously.)
- `number` → `schema.Float64Attribute` / `types.Float64` / `*float64` /
  `types.Float64Value(*v)`.
- `boolean` → `schema.BoolAttribute` / `types.Bool` / `*bool` /
  `types.BoolValue(*v)`.
- `string` + `format: date-time` → SDK getter returns `*time.Time`; schema
  `StringAttribute`, model `types.String`; RHS `types.StringValue(v.String())`.
  Never import/reference the `time` package — only call the method `v.String()`.
- `string` + `enum` (named SDK type) → `StringAttribute` / `types.String`; getter
  returns `*<EnumType>`; RHS `types.StringValue(string(*v))`.
- object → `schema.SingleNestedBlock`; array of object → `schema.ListNestedBlock`;
  each adds a nested struct to `.Models` and recurses via `nestedBody`.

**JSON:API envelope flattening**:
- Singular: response `data` → the item schema. Map `data.id` → top-level `id`;
  flatten `data.attributes.*` to **top-level** computed attributes; **drop** the
  `type` discriminator. In `updateState`, preamble `attributes := data.GetAttributes()`,
  then guard each field `if v, ok := attributes.Get<Field>Ok(); ok && v != nil { state.X = … }`,
  and `id` via `data.GetIdOk()`.
- The singular `updateState` parameter type is the item struct by value
  (`<pkg>.<ItemType>`); `resp.GetData()` and `items[0]` both yield it.

**Attributes**:
- All projected data attributes are `Computed: true`.
- `id` is emitted by the template, not by you — don't add it to `.Schema.Attributes`.
  (singular read-only → Required; search-only → `utils.ResourceIDAttribute()`;
  both → Optional+Computed; plural → `utils.ResourceIDAttribute()`, hash-derived.)
- Query/list filter parameters become `Optional: true` attributes and feed
  `optionalParams` (singular `.Search.Filters` / plural `.Read.Filters`).
- Pagination: set the paginated branch when the list operation has an
  `x-pagination` extension; otherwise use the plain `resp.Data...` branch.
