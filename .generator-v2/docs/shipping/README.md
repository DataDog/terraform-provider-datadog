---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
---

# Shipping a generated pull request

**Who this is for:** a provider maintainer holding a generated pull request, or an
API owner who has been pointed at one and asked to finish it.

A generated pull request opens as a **draft** and is not evidence that the artifact
works. Generation proves the operation was representable and the code was written.
It does not prove the generated call does the right thing against the API.

What the pull request contains:

- the generated resource or data-source implementation
- its registration (`resources_generated.go` or `datasources_generated.go`)
- a Terraform example, used by documentation generation
- an acceptance-test **scaffold**, with TODOs
- generator diagnostics for the run
- a link back to the source API-spec pull request

## What you owe it before merge

In order, because each step depends on the last:

1. **[Complete the acceptance test](acceptance-tests.md).** The scaffold asserts
   little more than that an ID is set. This is the bulk of the work.
2. **[Record and replay its cassette](cassettes.md)** against the Frog organisation,
   and read the recording before committing it.
3. **[Work the review checklist](review-checklist.md)** — schema, request, state
   mapping, diagnostics.
4. Review the example. Replace it if the pull request flags it as a placeholder.
5. Take it out of draft, and merge it through the normal provider workflow.

## What CI does and does not prove

| Check | What it establishes |
|---|---|
| `make docs` | The example can produce a documentation page |
| `make build` | The provider compiles and formatting passes |
| `make test` | Provider unit tests pass |
| `RECORD=false make testacc` | The test still agrees with its committed cassette |
| Live integration workflow | Selected acceptance tests pass against the current API |

Ordinary pull-request CI **replays** committed cassettes across the Terraform and
OpenTofu matrix. It never records or refreshes them. So CI can be green on a test
that proves very little — a cassette only covers the interactions recorded during
that one run.

A green replay also does not prove the API still behaves that way. That is what the
live run in [cassettes.md](cassettes.md) is for.

## Quality gates to run locally

Before committing:

```sh
make fmtcheck
make test
```

Before pushing:

```sh
make docs
make check-docs
make vet
make errcheck
```

Do not run raw `go test` for acceptance tests. The Makefile targets supply the test
runner and the acceptance-test configuration; bare `go test` will not.

## If something looks wrong rather than unfinished

Read the run report's diagnostics first — a dropped field or an unrepresentable
shape is usually reported rather than silently mishandled. See
[reference/diagnostics.md](../reference/diagnostics.md).
