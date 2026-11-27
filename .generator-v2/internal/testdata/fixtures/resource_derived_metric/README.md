# Required caller-assigned identity

Reduced derived-metric CRUD spec from the TF-426 assessment. The authoritative
create schema requires `data.id`; no production resource is registered by this fixture.

The golden test checks generated output. The identity integration test reuses
this API contract under a synthetic Terraform resource name; see
`../resource_identity/README.md`.
