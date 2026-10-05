---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291536523"
  confluence_id: "7291667269"
---

# Stage 2 — Model

**Who this is for:** you are changing how an operation becomes a Terraform artifact.

`model.BuildArtifact` projects one tracked operation into a `model.Artifact`.
Package: `internal/model/`.

## The Artifact

One type serves both kinds. It carries:

- the Terraform name, artifact kind, and top-level description
- cardinality (singular or plural; ignored for resources)
- an `AttributeTree`
- `LifecycleBindings` — the SDK calls per role
- the destination source path and generated-file ownership
- non-fatal `Diagnostic`s

`AttributeTree` mirrors the Terraform Plugin Framework schema. Each recursive
`Attribute` holds its Terraform type, generated Go type, collection element types,
description, flags, validators, nested children, component identity, and an optional
`oneOf` envelope. Maps, nested objects, lists, and supported unions stay structural
— nothing degrades to a dynamic value.

## Lifecycle roles

| Role | Used by |
|---|---|
| `Read` | Both. For a plural data source this is the **list** call. |
| `Search` | Singular data sources only — a list endpoint resolving exactly one record. |
| `Create` / `Update` / `Delete` | Resources. |

`lifecycle.go` builds these. A resource with no `update` marks all attributes
ForceNew — the missing-CRUD edge case.

## The resource schema merge

`merge.go` is the densest code in the package and the part most likely to surprise
you. A resource has up to three bodies — create, update, read — that must become one
Terraform schema.

Rules that are easy to break:

- **Validation widens, cosmetics follow the read.** The validator must accept
  everything the response can return, so enum members are unioned. Name,
  description, and `RefName` favour the read response.
- **`Sensitive` and `WriteOnlySecret` OR across roles.** If any role considers a
  field secret, the merged field is secret. Never AND these.
- **`RequestRequired` reads Create's list only.** Update-only required entries are
  deliberately ignored.
- **Per-role components are expected.** Create and update may reference different
  components and still merge into one field list; that is the normal case, not a
  conflict.
- A genuine conflict produces `resource schema merge conflict at "<path>"` and fails
  the artifact. Reconciled cosmetic differences produce a diagnostic instead.

## Framework type mapping

`framework_types.go` maps schema kinds to Framework attribute types. Formats are
deliberately ignored for integers and numbers: `int32` and `int64` both become
`Int64Attribute`, `double` becomes `Float64Attribute`. Do not reintroduce
format-sensitivity here without a reason — it was a conscious simplification.

Unrepresentable element or value kinds produce
`array element kind %q is not representable` / `map value kind %q is not
representable`, carrying the underlying `UnsupportedReason` through.

## Unsupported shapes stay explicit

`RefCycle`, `DepthExceeded`, and `UnsupportedReason` are all retained on the schema
rather than dropped, so a failure can say *why* and point at a position. Preserve
that when adding shapes: a silent omission is worse than a failed artifact.
