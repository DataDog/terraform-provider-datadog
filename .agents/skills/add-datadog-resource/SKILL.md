---
name: add-datadog-resource
description: Add a Datadog Terraform provider resource or data source. Route Datadog organization members to the OpenAPI generator skills in datadog-api-spec; guide external contributors through manual implementation using current Terraform Plugin Framework recommendations.
---

# Add a Datadog Terraform resource

Choose the contribution path before writing provider code. Read `DEVELOPMENT.md`
and `TESTING.md` at the provider checkout root, and inspect existing implementations
of the target API to avoid duplicate resources or accidental migrations.

## Establish contributor context

Use the contributor's stated affiliation or already verified context. Do not infer
DataDog GitHub organization membership from this repository's remote, a Git commit
email, a Graphite profile, or access to the public provider repository.

If membership is unknown and GitHub CLI authentication is available, follow any
applicable account-preflight instructions, verify the effective account with
`gh api user --jq .login`, and check that account's membership:

```sh
gh api user/memberships/orgs/DataDog --jq .state
```

An `active` membership selects the Datadog path. Missing authentication, insufficient
token scopes, a 404, or a pending membership leaves affiliation unresolved. Ask
whether the contributor is a DataDog organization member or an external contributor;
do not require external contributors to authenticate or obtain internal access.

## Datadog contributors: prefer the generator

Explain that Datadog contributors should declare Terraform artifacts in the source
OpenAPI specification and use tfgen rather than start with hand-written provider
code. The source repository is `datadog-api-spec`.

Locate an available API-spec checkout and read its agent instructions. Load the
relevant skill there, including only the references needed for the task. These
paths are relative to the **API-spec checkout**, not the provider checkout:

| Task | Skill |
| --- | --- |
| Declare a new resource or data source | `terraform-generator-annotate` |
| Try generation and assess support or replacement compatibility | `terraform-generator-assess` |
| Investigate failed or unexpected generation | `terraform-generator-diagnose` |
| Test generated artifacts and finish acceptance coverage | `terraform-generator-test` |

Each skill lives at `.agents/skills/<skill-name>/SKILL.md`. Refer to internal
skills by name and checkout-relative path without publishing internal repository
or documentation URLs. If no checkout is available, ask the contributor to locate
one or make the relevant skill instructions available through an authorized source.
If the instructions remain unavailable, identify the missing input or access
blocker; do not silently switch a Datadog contributor to manual implementation.

Keep annotation, CLI, and pipeline instructions in their maintained sources.
Provider-side references are `.generator-v2/docs/README.md` and
`.generator-v2/docs/reference/scope.md`. Fix source annotations or generator support
rather than hand-edit emitted files. If a demonstrated generator limitation makes
manual implementation necessary, explain the limitation and agree on that path
with the contributor before proceeding. Respect an explicitly requested manual path.

## External contributors: implement manually

Proceed in this public repository without requiring access to `datadog-api-spec`,
internal tooling, or the generator. Before designing the implementation, consult
the current official [Plugin Framework documentation](https://developer.hashicorp.com/terraform/plugin/framework),
[provider best practices](https://developer.hashicorp.com/terraform/plugin/best-practices),
and the [resource lifecycle guidance](https://developer.hashicorp.com/terraform/plugin/framework/resources)
or [data source guidance](https://developer.hashicorp.com/terraform/plugin/framework/data-sources)
for the target kind. Check recommendations against the dependencies pinned in
`go.mod` and the Terraform versions supported by this repository. If current docs
are unavailable, state that limitation instead of claiming the guidance is current.

- Search `datadog/fwprovider/` for comparable hand-written implementations and
  shared helpers. Use Terraform Plugin Framework for new resources and data
  sources; legacy SDKv2 implementations are not templates for new work.
- Use the Datadog Go API client and the provider's configured clients/authentication.
  If the SDK lacks the operation, identify the SDK update needed; do not substitute
  raw HTTP calls.
- Add descriptions for every schema attribute. Follow `DEVELOPMENT.md` for nested
  attributes and secret handling, and the current Framework guidance for validators,
  defaults, and replacement rules. Use framework value types to preserve null and
  unknown values; sensitivity alone does not keep a secret out of state.
- Implement the API's lifecycle and import semantics, including remote deletion
  during refresh and idempotent deletion. Preserve configured values consistently
  with Terraform's plan/state rules. For data sources, implement the appropriate
  lookup/list semantics instead of resource CRUD and import.
- Register the constructor in the appropriate hand-written provider registry,
  following `datadog/fwprovider/framework_provider.go`. Leave generated registries
  and files untouched. Add HCL examples under `examples/` and generate documentation
  with `make docs`; do not edit generated `docs/` files by hand.
- Add meaningful acceptance coverage for lifecycle, import, refresh/no-change
  plans, and relevant edge cases; for data sources, cover lookup/list behavior.
  Follow `TESTING.md` and repository quality gates. Run Go tests through Makefile
  targets, never raw `go test`. Use `RECORD=false` for existing cassette replay;
  recording and live modes call real APIs and require a suitable test organization.

Report the selected contribution path, changed files, checks actually run, and
any unverified behavior or blockers. Generating code or passing a build alone
does not establish runtime correctness.
