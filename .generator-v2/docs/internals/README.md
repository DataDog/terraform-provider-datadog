---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
---

# tfgen internals

**Who this is for:** you are changing the generator, not using it.

tfgen is a single Go module under `.generator-v2/`, separate from the provider
module. `make tfgen-build` produces `bin/tfgen`; `make tfgen-test` runs its suite.

## The four stages

`generate` orchestrates four responsibilities, in order. Each has its own page.

```
OpenAPI document
  │
  ├─ 1. PARSE ────────────────────── internal/parser/     → model.Spec
  │     resolve refs, normalize schemas, decode annotations
  │
  ├─ 2. MODEL ────────────────────── internal/model/      → model.Artifact
  │     project one annotated operation into a Terraform-facing artifact
  │
  ├─ 3. BIND ─────────── internal/sdkbind{,ing}/          → model.SDKCall
  │     resolve the Datadog Go SDK call the artifact must make
  │
  └─ 4. EMIT ─────────────────────── internal/emit/       → .go / .tf / _test.go
        render views through templates, wire registration, reconcile, report
```

Orchestration lives in `internal/cli/generate.go`; retirement and reconciliation in
`internal/cli/reconcile.go`.

| Stage | Page | Package |
|---|---|---|
| Parse | [parser.md](parser.md) | `internal/parser/` |
| Model | [model.md](model.md) | `internal/model/` |
| Bind | [sdk-binding.md](sdk-binding.md) | `internal/sdkbind/`, `internal/sdkbinding/` |
| Emit | [emit.md](emit.md) | `internal/emit/` |

Adding support for something new: [extending.md](extending.md).

## Data flow

```
Spec
└── Operation
    ├── TrackingFieldMetadata        the decoded x-datadog-tf-generator
    ├── Schema (request + response)  normalized, recursive
    └── SDKOperationBinding          derived from OpenAPI metadata

Operation
└── Artifact                         one per tracked annotation
    ├── AttributeTree                mirrors the Framework schema
    ├── LifecycleBindings
    │   └── SDKCall                  per lifecycle role
    └── []Diagnostic                 non-fatal

RunReport                            artifacts, skipped operations, summary
```

There is **one** `Artifact` type, not separate resource and data-source structs.
Resources populate the Create/Update/Delete lifecycle roles; data sources populate
Read and optionally Search. Both are the same shape in the model, which is why
resource support reused most of the data-source pipeline.

## Principles that the code actually holds to

Worth knowing before you change anything, because they are load-bearing:

- **Fail rather than guess.** An unrepresentable shape fails its artifact with a
  diagnostic. There is no dynamic-value escape hatch, and adding one would change
  the generator's contract.
- **Determinism.** Operations are sorted; output is compared byte for byte against
  goldens in `internal/snapshots/`. Anything order-dependent or time-dependent is a
  bug — the run report's IDs and timestamps are the deliberate exception.
- **Local failure.** One bad artifact must not take down the batch. Only spec-loading
  failures abort the run, and that blast radius is why annotation validation is
  strict.
- **Write only on change.** The emitter compares before writing, which is what makes
  `--check` and `make tfgen-docs-check` possible.
- **Never clobber human work.** Existing generated test files are preserved. Files
  without the tfgen marker are never overwritten without `overwrites`.

## Contracts

`internal/contracts/` holds the machine-readable contracts, embedded so the binary
never depends on its working directory:

- `tracking-field.schema.json` — the annotation. `parser.DecodeTracking` compiles it
  once and validates every occurrence.

It is rendered into [reference/annotation.md](../reference/annotation.md) by
`cmd/tfgen-docs`. Change the schema, run `make tfgen-docs`, commit both.

The `--report` output has **no** contract here yet, which is why
[reference/diagnostics.md](../reference/diagnostics.md) describes its statuses by
hand. Adding a `run-report.schema.json` would let that page be generated too —
`cmd/tfgen-docs` is already generic over any schema in this package.
