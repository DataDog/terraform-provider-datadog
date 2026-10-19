---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291536523"
  confluence_id: "7290881599"
---

# Stage 4: Emit

**Who this is for:** you are changing the generated output.

`internal/emit/` turns an `Artifact` into template-facing views and renders them.
This is where golden snapshots bite, so expect `make tfgen-update-goldens` in any
change here.

## View, then template

`builder.go` and `view.go` build a flat, template-facing view; `templates/` holds the
Go templates; `render.go` executes them; `writer.go` writes only on change.

```
templates/
├── resource.go.tmpl            full CRUD
├── data_source_singular.go.tmpl
├── data_source_plural.go.tmpl
├── data_source_common.go.tmpl  shared blocks
└── data_source_test.go.tmpl    acceptance-test scaffold
```

The split exists because the three artifact shapes differ more in their read/write
logic than in their schema rendering. Put shared logic in the view, not in template
conditionals.

## Envelope flattening

`flattenEnvelope` turns the Datadog v2 JSON:API envelope into a flat Terraform
schema:

- `data.id` → top-level `id`
- `data.attributes.*` → top-level attributes
- `data.type` omitted
- relationships, sideloaded data, request metadata omitted
- `created_at`, `created_by`, `modified_at` omitted, each with an `info` diagnostic

It is also where `id_strategy` is enforced: anything other than `data.id` fails
here with `id_strategy %q is not yet supported (only data.id)`. For a plural data
source the same transformation applies per item.

Every omission emits a diagnostic. Keep that property: the report is how a reviewer
learns a field is gone.

## State mapping

Generated mapping uses the SDK's presence-aware getters, so an omitted API field
becomes a Terraform **null** rather than an empty string or zero. Lists, maps, nested
objects, enums, UUIDs, and dates each get type-specific mapping. A `oneOf` mapper
verifies that exactly one alternative matched and produces a diagnostic for missing,
ambiguous, or unparsed unions instead of partial state.

## Naming

`modelname.go` derives generated Go model names using component identity, so two
distinct OpenAPI components cannot collide on one struct name. `accessors.go`
resolves provider API accessors; `apipath.go` handles path construction;
`unstable_operations.go` handles operations the SDK marks unstable.

## Registration and wiring

`registration.go` and `resource_registration.go` maintain:

- `datasources_generated.go` / `resources_generated.go`: a **union merge**, sorted,
  so a scoped `--include` run never drops artifacts it did not regenerate
- `framework_provider.go`: removal of a hand-written constructor when the annotation
  declares `overwrites`
- `datadog/tests/provider_test.go`: the `testFiles2EndpointTags` map entry for a
  generated test

Because the registry files are sorted and merged, two concurrent artifact pull
requests both touch them; the first to merge shifts the base and the second needs a
trivial rebase. The resolution is deterministic: re-run the scoped emit.

## What is preserved

- Existing generated **test** files are never overwritten, so a completed acceptance
  test survives regeneration.
- Examples are create-only: tfgen writes `data-source.tf` only when absent, so a
  maintainer's improved example is never clobbered.
- A file without the tfgen generated-code marker is never overwritten without
  `overwrites`.

Markdown documentation is not produced here. It comes from `make docs` via
`tfplugindocs`, which consumes the generated example.

## Goldens

`internal/snapshots/*.golden` are compared byte for byte by the emit tests. An
intentional change shows up as a reviewable snapshot diff; refresh with
`make tfgen-update-goldens` and read the diff before committing it.
