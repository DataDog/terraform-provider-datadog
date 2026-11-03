# Synthetic caller-assigned identity resource

`TestGenerateCallerAssignedIdentity` creates a temporary OpenAPI fixture named
`tfgen_identity_fixture` from the derived-metric API contract. The Terraform
resource name is changed before generation, so the generated Go declarations
and Terraform type are distinct from the production derived-metric resource.
The generated resource and this probe are added through a temporary Go overlay;
no existing provider resource is replaced or modified.

The API operations and schemas retain their real Datadog SDK bindings. The
probe sends requests to a local HTTP server, without Datadog credentials.
A fully fictional API would also require a synthetic SDK.

The test generates twice and checks stability, then executes
`identity_test.go.txt` against the synthetic resource. It covers required name
configuration, computed API identity, canonical-ID routing, PUT bodies,
replacement modifiers, CRUD state, and Terraform import/refresh/adoption.

A negative subtest changes the generated create request to use the unknown
computed ID instead of the configured name. It requires the probe to fail with
the empty POST ID diagnostic, so a compilation or environment failure cannot
accidentally satisfy the negative test.
Only one direct CRUD subtest runs against the mutated source; the negative test
does not repeat generation or Terraform import/adoption. The changed source
still requires a separate Go test invocation to compile and execute it.

Run `make tfgen-test-integration` from the repository root.
