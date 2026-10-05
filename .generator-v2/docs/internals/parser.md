---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
---

# Stage 1 — Parse

**Who this is for:** you are changing how tfgen reads OpenAPI.

`parser.LoadSpec` reads the document and produces a `model.Spec`. Package:
`internal/parser/`.

## Order of work

1. Read and hash the specification (`spec_hash` in the report is a SHA-256 of it).
2. Build the OpenAPI v3 model, via `pb33f/libopenapi`.
3. Detect reference cycles and paths exceeding `--max-depth` (`cycles.go`).
4. Decode `x-datadog-tf-generator` annotations (`tracking.go`).
5. Sort operations deterministically by path and method.
6. Reject duplicate `(artifact_kind, artifact_name)` claims (`duplicates.go`).
7. Normalize request and response schemas for tracked operation groups
   (`schema.go`).

Steps 3 and 4 abort the whole run on failure; later failures are per-artifact. That
asymmetry is the single most important thing to preserve when changing this package.

## Annotation decoding

`DecodeTracking` validates every occurrence against the embedded
`tracking-field.schema.json` using `santhosh-tekuri/jsonschema/v6`. The schema sets
`additionalProperties: false`, so an unknown field is a hard error rather than a
warning.

The extension key is configurable via `--tracking-field`, which exists for fixtures
rather than for production use.

**Schema-node annotations take a different path.** A `sensitive` marker on a Schema
Object is read by `isSensitive` through a narrow struct decode, which does *not* run
the JSON-Schema validation that operation annotations get. That is why a bare
`{sensitive: true}` works even though the flat contract nominally requires
`artifact_kind` and `artifact_name` on every occurrence. If you unify these paths, a
lot of existing specs will start failing validation — intentionally or not.

A malformed `sensitive` annotation counts as absent, leaving `writeOnly` / `x-secret`
inference in force. That direction is deliberate: it fails safe, so a typo cannot
expose a credential in a plan.

## Normalization

`schema.go` converts OpenAPI composition into the generator's recursive `Schema`:

- **`allOf`** — compatible branches merged; there is no retained `allOf` in later
  stages. Conflicting intersections become unsupported nodes at the affected
  position, not run failures. Watch `isAnnotationOnlySchema`: a `$ref` carrying only
  a `readOnly` sibling must not be treated as a composition, and getting that wrong
  produced a real merge bug.
- **`oneOf`** — becomes a typed `OneOfSpec` with discriminator info and alternatives.
- **`additionalProperties`** — recognized as an object map.
- **`anyOf`** — retained as unsupported, with a reason. First match wins in the
  unsupported classification, since a node can satisfy several conditions at once.

Descriptions, formats, enums, component names, and sensitivity markings are all
preserved for later stages. Component identity in particular matters: the emitter
uses it so two distinct OpenAPI components never collapse into one Go struct name.
