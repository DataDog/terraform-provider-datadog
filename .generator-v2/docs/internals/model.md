---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291536523"
  confluence_id: "7291667269"
---

# Stage 2: Model

**Who this is for:** you are changing how an operation becomes a Terraform artifact.

`model.BuildArtifact` projects one tracked operation into a `model.Artifact`.
Package: `internal/model/`.

## The Artifact

One type serves both kinds. It carries:

- the Terraform name, artifact kind, and top-level description
- cardinality (singular or plural; ignored for resources)
- an `AttributeTree`
- `LifecycleBindings`: the SDK calls per role
- the destination source path and generated-file ownership
- non-fatal `Diagnostic`s

`AttributeTree` mirrors the Terraform Plugin Framework schema. Each recursive
`Attribute` holds its Terraform type, generated Go type, collection element types,
description, flags, validators, nested children, component identity, and an optional
`oneOf` envelope. Maps, nested objects, lists, and supported unions stay structural
Nothing degrades to a dynamic value.

## Lifecycle roles

| Role | Used by |
|---|---|
| `Read` | Both. For a plural data source this is the **list** call. |
| `Search` | Singular data sources only: a list endpoint resolving exactly one record. |
| `Create` / `Update` / `Delete` | Resources. |

`lifecycle.go` builds these. When a resource has no `update`, configurable inputs
require replacement on change; computed-only response fields do not.

## The resource schema merge

`merge.go` is the densest code in the package and the part most likely to surprise
you. A resource has up to three bodies (create, update, read) that must become one
Terraform schema.

Rules that are easy to break:

- **Validation widens, cosmetics follow the read.** The validator must accept
  everything the response can return, so enum members are unioned. Name,
  description, and `RefName` favour the read response.
- **Sensitivity and write-only handling are distinct.** `Sensitive` ORs across
  roles. `WriteOnlySecret` comes from writable request roles; a response-only
  marker must not enable write-only handling. See
  [secret handling](../reference/scope.md#secrets-and-sensitivity).
- **`RequestRequired` reads Create's list only.** Update-only required entries are
  deliberately ignored.
- **Provenance retains create-only inputs.** `SchemaProvenance.CreateOnly` records
  writable Create presence without writable Update presence. Read-only request
  properties are excluded before merging. `schema.go` uses this bit for typed
  `RequiresReplace` modifiers while preserving mutable siblings and computed-only
  fields. See the [lifecycle rules](../reference/scope.md#schema-shapes).
- Immutable objects and object lists use structural replacement modifiers on
  their parents and value replacement modifiers on their writable descendants.
  This avoids comparing computed descendants before their plan modifiers run.
- **Per-role components are expected.** Create and update may reference different
  components and still merge into one field list; that is the normal case, not a
  conflict.
- A genuine conflict produces `resource schema merge conflict at "<path>"` and fails
  the artifact. Reconciled cosmetic differences produce a diagnostic instead.

## Framework type mapping

`framework_types.go` maps schema kinds to Framework attribute types. Formats are
deliberately ignored for integers and numbers: `int32` and `int64` both become
`Int64Attribute`, `double` becomes `Float64Attribute`. Do not reintroduce
format-sensitivity here without a reason; it was a conscious simplification.

Unrepresentable element or value kinds produce
`array element kind %q is not representable` / `map value kind %q is not
representable`, and a nested collection's inner element produces
`collection element kind %q has no framework element type`. Each carries the
underlying `UnsupportedReason` through. The tree builder wraps these in
`AttributeError`, which prefixes the attribute path:
`model: cannot build attribute at "<path>": map value kind ...`.

For nested collection failures, `ElementType` retains a relative element path in
an internal error wrapper. The tree builder adds the outer collection's marker
and attribute path to populate `AttributeError.ElementPath`. The message includes
`element at "<element path>"`, using `[]` for list items and `{}` for map values.
`AttributeError.Path` continues to identify the attribute, and `Unwrap` preserves
the mapping failure for `errors.As` and other error inspection. Single-level
failures leave `ElementPath` empty and retain their previous message format.

## Unsupported shapes stay explicit

`RefCycle`, `DepthExceeded`, and `UnsupportedReason` are all retained on the schema
rather than dropped, so a failure can say *why* and point at a position. Preserve
that when adding shapes: a silent omission is worse than a failed artifact.
