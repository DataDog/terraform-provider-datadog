---
ddoc:
  confluence_id: "6733104627"
  confluence_space: "API"
  confluence_parent: "4784881958"
---

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

## Mirroring to Confluence

These files are the source of truth for the Confluence docs in the `API` space. ddoc
has no repository-level config; it is configured per file, in the leading YAML block
each page carries.

The mapping is deliberately simple:

- **`README.md` carries `confluence_id: "6733104627"`** and so replaces the existing
  [Terraform Generator v2](https://datadoghq.atlassian.net/wiki/spaces/API/pages/6733104627)
  page in place. That page stays the entry point and keeps its URL and inbound links.
- **Every other page is new** and carries `confluence_parent: "6733104627"`, so it is
  created under that entry point rather than at the space root.
- **The old pages it supersedes are deleted**, not updated — the structure changed too
  much for a page-for-page mapping, and three of them split into several.

> **ddoc never deletes a Confluence page.** Removing Markdown leaves its page in
> place, so the superseded pages have to be deleted by hand. Until that happens they
> coexist with the new tree, which is the one state worth avoiding — two sets of docs
> saying different things is how this drifted in the first place.

Pages to delete once the new tree is published: v2 — Overview, Quick Start, Technical
Overview, Pipeline in Depth, Generation in Depth, Testing, Internal data models, CLI
contract, OAS extension contract, Road to Resources / Future Plans, and the
`Contracts` folder that held the last two.

Publishing is automatic: the org-scoped dd-octo-sts policy in `DataDog/.github`
(`ddoc-sync-consumer-read-repos`, `repositories: []`) already authorizes this
repository, and `ddoc-sync-consumer` publishes on every push to the default branch.
So **merging a change to any of these files republishes its page.** Nothing happens on
a feature branch.

Before the first publish, run a dry run and read the plan:

```sh
ddoc sync --dry-run .generator-v2/docs
```

The Confluence tree mirrors this directory structure: each directory's `README.md`
is the page its siblings hang under, and `pipeline.md` sits at the top alongside them.

Two ddoc behaviours worth knowing:

- Only a file named exactly **`index.md`** parents its siblings automatically, and
  inheritance is not recursive. `README.md` is not special to ddoc, so each page names
  its parent explicitly in `confluence_parent`. Adding a page means pointing it at its
  directory's index page ID.
- The sync service never writes resolved IDs back to GitHub — only the local CLI does.
  Every page here already records its `confluence_id`, so updates target pages by ID
  rather than resolving them by title. That matters: "Reference" was already taken in
  this space, which is why the reference index is titled "tfgen reference".
