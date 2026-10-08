---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291536637"
  confluence_id: "7291536545"
---

# Scope: what tfgen supports today

**Who this is for:** anyone asking "can tfgen generate this?", before annotating,
before reviewing, before filing a bug.

This is the only page that states what the generator can and cannot do. Other pages
link here rather than summarising, so there is exactly one place to change when
support changes, and exactly one place to check.

> **Changing what tfgen supports? Edit this page in the same pull request.** The
> previous documentation set stated the supported artifact kinds on seven separate
> pages, so when resource generation landed, all seven became wrong at once.

## Artifact kinds

| Kind | Status |
|---|---|
| **Resource** (full CRUD) | **Supported and shipping.** Generation landed in #4227. Four generated resources are in the provider today: the Databricks, Elastic Cloud, Snowflake, and Twilio integration-account resources. |
| **Data source**, singular | **Supported by the generator; none currently shipped.** tfgen emits them and the golden snapshots cover them, but `generatedDatasources` in the provider is an empty slice; no generated data source has been adopted yet. |
| **Data source**, plural | Same: emitted and covered, none shipped. |

If you are generating a data source, you are the first. That is workable. See
[generating/data-sources.md](../generating/data-sources.md), but expect to be the
one who discovers the rough edges, and budget time for the acceptance test.

## API surface

- **Datadog v2 only.** There is no V1 spec slice and no V1 support.
- The endpoint and its models must exist in the **Go SDK version pinned by the
  provider**. OpenAPI-derived binding is authoritative and generation can succeed
  before the SDK catches up, but the provider will not compile. SDK update and
  provider build are separate gates.

## Terraform identity

Only `data.id` is supported for data-source identity. The annotation schema accepts
all four `id_strategy` values, and the parser carries them, but emission rejects the
others with `id_strategy %q is not yet supported (only data.id)`.

## Schema shapes

**Supported:**

- scalars, nested objects, lists, and sets
- object maps, via `additionalProperties`
- `allOf`: compatible branches are merged during normalization; conflicting
  intersections are recorded as unsupported at the affected node
- `oneOf`: discriminated unions, at their own position or inside a list

**Not supported:**

- `anyOf`: retained as an explicit unsupported node with a reason, never guessed at
- a `oneOf` used as a **map value**
- reference cycles, and expansion deeper than `--max-depth` (default 20): both stop
  specification loading rather than failing a single artifact

An unsupported shape fails its artifact with a diagnostic instead of emitting a
partial schema. Independent artifacts in the same run are unaffected, unless the
failure prevented the spec from loading at all.

## Secrets and sensitivity

- An OpenAPI property marked `writeOnly: true` becomes Terraform write-only
  handling: the generated schema replaces it with three attributes — `<attr>_wo`
  and `<attr>_wo_version` for Terraform 1.11+, plus a stateful plaintext
  `<attr>` as the fallback for older Terraform. The two halves are mutually
  exclusive, and the plaintext one carries `PreferWriteOnlyAttribute`, so a
  1.11+ user is steered to the write-only path while a pre-1.11 user can still
  apply the resource.
- Requiredness lands on the pair, not on either half: a secret the API requires
  on create becomes `ExactlyOneOf(<attr>, <attr>_wo)`, and an optional one
  becomes `ConflictsWith`. Neither attribute is ever `Required` on its own,
  since either one alone satisfies the API.
- `x-secret: true` or `writeOnly: true` default an attribute to Terraform-sensitive.
  An explicit `sensitive: false` on the same schema node overrides that inference.
- Sensitive and `x-secret` redact from display; they do **not** keep values out of
  Terraform state. Only write-only handling does that.

Generated and hand-written resources share the same three-attribute pattern via
`datadog/internal/fwutils`. Generated resources pass
`WriteOnlySecretModeDual`, which differs from hand-written
`WriteOnlySecretModeLegacy` only in honouring `Required`: Legacy predates it and
always forces `ExactlyOneOf`.

## What a green run does not prove

A successful `tfgen generate` means the operation was representable and the code was
written. It does not mean:

- the provider compiles; that is `make build`
- the generated call works against the API; that is a recorded cassette
- the acceptance test is meaningful; the generated test is a **scaffold** with TODOs
- the API still behaves as recorded; that is live validation

See [shipping/](../shipping/README.md).

## Not supported, and sometimes assumed to be

Two of these are listed because earlier documentation described them as working.

| Thing | Reality |
|---|---|
| `ignore`, omitting response attributes via the annotation | **Does not exist.** It was never merged to the default branch. The annotation schema sets `additionalProperties: false`, so passing `ignore` does not degrade gracefully: it fails validation at parse time and aborts the whole run with no report written. Note the transformer still accepts the field, so it passes upstream validation and only fails here. |
| `tfgen verify` | Registered, performs no checks, returns success. A passing `verify` is evidence of nothing. |
| Go hooks / `--hooks-root` | Accepted and ignored. Hook discovery is not implemented. |
| `--quiet` | Accepted and ignored. |
| Resource retirement via `overwrites` | Retiring a generated artifact that replaced a hand-written one does not resurrect the original. |

## Known gaps worth knowing before you start

- Generated acceptance tests are scaffolds requiring human completion. This is
  currently the main cost of adopting a generated artifact.
- Artifact processing in the delivery pipeline is sequential.
- Retirement is fail-closed: a file without the tfgen marker, or one with a recorded
  cassette, is never deleted automatically.
