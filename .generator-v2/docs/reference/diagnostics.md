---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291536637"
  confluence_id: "7291470939"
---

# Interpreting a failed run

**Who this is for:** your run failed, or produced something you did not expect, and
you want to know why before changing anything.

`tfgen generate --report <path>` writes a JSON report; `-` writes it to stdout. It
carries run metadata (`run_id`, `generator_version`, a SHA-256 `spec_hash`, start and
finish times), one entry per emitted artifact file, a list of `skipped_operations`,
and summary counts. Run IDs and timestamps change every invocation, so a whole report
is never byte-for-byte reproducible; compare fields, not files.

> **There is no generated reference for the report yet,** because the report has no
> machine-readable contract in `internal/contracts/`. The statuses below are
> hand-maintained and may lag the code. When a schema lands, add a generated page
> rather than growing a field table here; see house rule 2 in
> [../README.md](../README.md).

Artifact entries carry a `name`, `kind`, `status`, output `path`, and any
`diagnostics`. Entries for declared artifacts also carry `tag`, the name of the
owning registry's OpenAPI tag. This can differ from the backing endpoint's tag.
Generation scripts use the resolved name and tag from this report for PR service
prefixes. Retired artifacts have no declaration and omit `tag`; their prefix
falls back to `core`. The statuses are `created`, `updated`, `unchanged`, `skipped`,
`failed`, `retired`, `retire_blocked`, and `registration_retired`. Note that each
generated *file* is its own entry, so the summary counts report entries rather than
logical artifacts. One data source producing a source file and an example counts as
two.

The rest of this page is the part that could not be generated from anything: what a
given outcome usually means.

> **This catalogue is partial by construction.** Diagnostic text is built inline
> across `internal/parser/` and `internal/model/` rather than from a central table,
> so there is no complete list to generate from. The entries below are the failures
> people actually hit. If you hit one that is not here, add it.

## First: did the run produce a report at all?

This distinction matters more than the exit code.

| Symptom | Meaning |
|---|---|
| A report was written, some artifacts `failed` | Per-artifact failure. Other artifacts were unaffected. Read their diagnostics. |
| **No report at all**, exit `1` | The run aborted before artifacts were evaluated, almost always spec loading. **Every artifact in the run was lost, including other teams'.** |
| Exit `3` | `--check` only: generated files or wiring would change. Not an error; regenerate and commit. |

## Whole-run aborts

These stop specification loading, so nothing is generated.

**`tag "<tag>": parser: invalid family registry ...`**: a tag registry does not
validate against the schema, or a tag/family is declared more than once. The commonest cause is a field that is not in the
contract, because `additionalProperties` is `false`:

```
- at '': additional properties 'ignore' not allowed
```

`ignore` specifically is documented in older material but was never merged. See the
unsupported table in [scope.md](scope.md). Fix the annotation in the spec; the
accepted fields are in [annotation.md](annotation.md).

**Reference cycle / depth exceeded**: a `$ref` re-enters its own path, or expansion
passed `--max-depth` (default 20). Raising the limit is rarely the answer; a cycle
usually indicates a schema that cannot be projected into a Terraform schema at all.

**Duplicate artifact claim**: two family blocks claim the same
`(artifact_kind, artifact_name)`. Names must be unique within a kind; a resource and
a data source may share one.

## Per-artifact failures

**`group.<role> names operationId "<id>", which the document does not declare`**:
an artifact's primary operation is missing. The diagnostic identifies the tag,
family, and block. Fix the operation ID in that block's group. Other artifacts
continue generating and the failed declaration appears in the report. Any
artifact failure prevents `--reconcile` from retiring orphans in that run.

**`<kind> kind "<k>" is not representable`**: a schema shape the model cannot
express. Check it against the supported-shapes list in [scope.md](scope.md); `anyOf`
and a `oneOf` used as a map value are the usual answers.

**`resource schema merge conflict at "<path>"`**: the create, update, and read
bodies disagree about that path in a way the merge will not reconcile. Most often
one side has the field as an object and another as something incompatible. Fix the
spec; do not work around it in generated code.

**`id_strategy "<s>" is not yet supported (only data.id)`**: exactly what it says.

**SDK binding failure**: the generated call cannot be resolved against the Go SDK.
Generation can still succeed, because OpenAPI-derived binding is authoritative and
the pinned SDK is only corroboration; the failure then surfaces at `make build`
instead. If the endpoint is newly added, the SDK bump has to land first.

**`refusing to overwrite <file>: it carries no Code generated by tfgen marker`**:
a hand-written artifact already occupies that path. Intentional takeover needs
`overwrites`; see [generating/resources.md](../generating/resources.md).

## Diagnostics that are not failures

| Severity | What to do |
|---|---|
| `info` | Read it. Dropped audit fields and envelope flattening are normal, but this is also how the generator tells you a field you may care about is gone. |
| `warning` | Investigate before merging. Generation continued, but something needs a human. An unexamined warning is the most common route for a real problem to reach merge. |
| `error` | The artifact failed. |

Legacy operation declarations and schema-level `sensitive` overrides produce
warnings on stderr and remain ignored. Move declarations to tag families; use
`writeOnly` or `x-secret` for sensitivity. Warnings name the operation or source
line without printing annotation values.

CLI-level warnings, such as the pinned SDK being unavailable for corroboration,
go to stderr rather than attaching to an artifact entry, so check the run log as
well as the report.

## Retirement outcomes

Retirement is fail-closed, so a refusal is the system working.

- `retire_blocked`: the file has no tfgen marker, or a recorded cassette shows the
  artifact was adopted. Nothing is deleted; a human decides.
- `registration_retired`: a stale registration whose implementation files are
  already gone.
