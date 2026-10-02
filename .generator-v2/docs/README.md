# tfgen — the Datadog Terraform provider generator

tfgen turns annotated Datadog v2 OpenAPI operations into Terraform provider code.
An operation carrying `x-datadog-tf-generator` becomes a generated resource or data
source, registered in the provider and shipped like any other.

**Start with the question you arrived with.**

| You are… | Go to | What you get |
|---|---|---|
| An API owner who wants a Terraform artifact for your endpoint | [generating/](generating/README.md) | How to annotate, and what reaches you afterwards |
| A maintainer holding a generated pull request | [shipping/](shipping/README.md) | What to finish, test, and check before merge |
| Looking at a failed or surprising run | [reference/diagnostics.md](reference/diagnostics.md) | Report statuses, diagnostics, what each failure means |
| Changing tfgen itself | [internals/](internals/README.md) | Architecture, the stages, and where to add a shape |

**The two pages worth knowing by name:**

- [reference/scope.md](reference/scope.md) — the single statement of what tfgen can
  do today. No other page restates it. If you are wondering whether something is
  supported, that is the only page to read.
- [reference/annotation.md](reference/annotation.md) — every accepted annotation
  field, generated from the schema the generator validates against.

## How it fits together

```
OpenAPI operation (datadog-api-spec)
  │  annotated with x-datadog-tf-generator
  ▼
openapi-transformer ─── validates the annotation, carries it into the compiled spec
  │
  ▼
tfgen generate ──────── parse → model → bind to the Go SDK → emit
  │
  ▼
generated provider code + example + acceptance-test scaffold + run report
  │
  ▼
a human finishes the test, records a cassette, reviews, merges
```

The last step is not optional and not automated. A successful generation run means
the code compiles and the schema is representable — not that the artifact works
against the API. See [shipping/](shipping/README.md).

## Reference

| Page | What it is |
|---|---|
| [reference/scope.md](reference/scope.md) | What is supported, what is not — the single source |
| [reference/annotation.md](reference/annotation.md) | `x-datadog-tf-generator` fields · generated |
| [reference/cli.md](reference/cli.md) | `tfgen` commands and flags · generated |
| [reference/diagnostics.md](reference/diagnostics.md) | Reading a failed run |
| [pipeline.md](pipeline.md) | What triggers generation, and who owns it |
| [decisions/](decisions/) | Why things are the way they are, dated |

## House rules for these docs

These exist because this documentation set drifted badly once: an annotation field
was documented for seven weeks after it failed to merge, the stated scope was wrong
on seven pages at once, and half a page described a workflow that had been deleted.
Each rule below is a direct response to one of those.

1. **One scope file.** Only [reference/scope.md](reference/scope.md) states what is
   supported. Everything else links to it. Never restate it — that is how it went
   wrong on seven pages simultaneously.
2. **Tables of flags, fields, statuses, or codes are generated.** If it cannot be
   derived from a source of truth, do not tabulate it; write a sentence instead.
   See `cmd/tfgen-docs`.
3. **Never re-document another repository's workflow.** Name the owner, link, stop.
4. **Links point at paths on the default branch.** Never a commit permalink, never a
   branch — citing branch state as shipped state is what made an unmerged feature
   look real.
5. **Unmerged work gets a [decision record](decisions/), not a reference page.**
   Reference describes what runs today.
6. **Every page names its reader in the first sentence.** It is also the cheapest
   test of whether the page belongs where it sits.
7. **Working notes do not live here.** Not as a peer of a reference page.

Generated pages carry a `DO NOT EDIT` header. Regenerate with:

```sh
make tfgen-docs        # rewrite them
make tfgen-docs-check  # fail if any is stale (exit 3)
```
