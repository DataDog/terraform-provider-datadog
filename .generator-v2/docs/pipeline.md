---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
---

# What triggers generation

**Who this is for:** you annotated an operation and no provider pull request
appeared, or you are trying to find the run log.

## Ownership, in one table

| Repository | Owns |
|---|---|
| `DataDog/datadog-api-spec` | The operations and their annotations; building the compiled Terraform spec; **triggering generation** |
| `DataDog/openapi-transformer` | Validating `x-datadog-tf-generator` and carrying it into the compiled spec |
| `DataDog/terraform-provider-datadog` | tfgen itself, the generated code, tests, examples, and documentation |

## What happens

1. A **ready, non-draft** API-spec pull request without `ci/skip`, where the
   Terraform generator is selected by the changed-generator calculation.
2. The API-spec workflow builds the compiled Terraform specification and runs
   `tfgen generate` against it, before the spec pull request merges.
3. One invocation evaluates **every** annotated operation in the compiled spec — not
   only the operation your pull request changed. This is why a single invalid
   annotation can abort the batch; see
   [reference/diagnostics.md](reference/diagnostics.md).
4. The generated changes are routed into per-artifact draft provider pull requests.
5. A human finishes each one: [shipping/](shipping/README.md).

## Where the fan-out runs

**Not in this repository.** The `tfgen-split.yml` workflow that used to split an
aggregate branch into per-artifact pull requests was removed in #4262, on the
grounds that it belongs to the upstream pipeline rather than to the provider's own
pull requests. The `tfgen split` command and `internal/split/` remain here and are
still how the attribution is computed.

> **Gap, stated rather than papered over.** This page deliberately does not
> re-document the fan-out, because this repository does not own it — see house rule 3
> in [README.md](README.md). At the time of writing, the authoritative description of
> where it now runs had not been written in the owning repository. If you are
> debugging fan-out, start from the API-spec pipeline configuration, and please
> replace this paragraph with a link once that documentation exists.

## No provider pull request appeared

Work down this list:

- The API-spec pull request is ready for review, not a draft.
- It does not carry `ci/skip`.
- `x-datadog-generators.generators.terraform` is `true`.
- The `x-datadog-tf-generator` annotation is valid — see
  [reference/annotation.md](reference/annotation.md).
- Every `operationId` named under `group` exists and is spelled exactly.
- The endpoint and its models exist in the Go SDK version the provider pins.
- The generation report has no parser, representability, SDK-binding, or emission
  failure for your artifact — and no whole-run abort caused by *someone else's*
  annotation.

## Stale artefact worth cleaning up

`.github/chainguard/self.generator.tfgen-split.sts.yaml` still grants
`contents: write` and `pull_requests: write` to the removed `tfgen-split.yml`
workflow ref, and a comment in `.github/workflows/test.yml` still refers to that
workflow creating `generate/*` and `retire/*` branches. Neither affects generation;
both should go when someone is in there.
