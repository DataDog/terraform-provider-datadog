---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
  confluence_id: "7291536637"
---

# tfgen reference

**Who this is for:** you need a precise answer about a flag, a field, a status, or a
failure — not a walkthrough.

| Page | What it is | Maintained by |
|---|---|---|
| [scope.md](scope.md) | What tfgen supports today. The single statement; no other page restates it. | hand |
| [annotation.md](annotation.md) | Every accepted `x-datadog-tf-generator` field. | generated from `internal/contracts/tracking-field.schema.json` |
| [cli.md](cli.md) | `tfgen` commands, flags, and exit codes. | generated from the cobra command tree |
| [diagnostics.md](diagnostics.md) | Reading a failed run: report structure, statuses, what each failure means. | hand |

The two generated pages carry a `DO NOT EDIT` header. Change their source of truth and
run `make tfgen-docs`; `make tfgen-docs-check` fails while either is stale. A field
absent from a schema therefore cannot appear in its reference page — which is the
specific failure this directory exists to prevent.
