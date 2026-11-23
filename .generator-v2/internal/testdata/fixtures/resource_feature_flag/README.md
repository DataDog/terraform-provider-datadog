# Explicit feature flag ID parameter

`openapi.yaml` is a reduced snapshot of the Feature Flags API contract used by
the provider's pinned Datadog Go SDK. It retains real Create, Get, Update and
Archive operations, UUID path/response IDs, and a small set of real attributes.
The Terraform artifact has a synthetic name to avoid replacing a production
resource. Its resource annotation selects `id_parameter: feature_flag_id`.

`make tfgen-update-goldens` regenerates `out/`; `make tfgen-test` verifies that
the committed output matches fresh generation. The golden has one computed
`id`, no `feature_flag_id` input, a single-ID import, and UUID conversions for
Get, Update and the `/feature-flags/{feature_flag_id}/archive` action route.
Existing fixtures omit `id_parameter` and retain their generation behavior.

`make tfgen-test-integration` runs `TestGenerateFeatureFlagIDParameter`. It
generates twice, requires identical output, and overlays the fresh resource
and `id_parameter_test.go.txt` into the provider package. The generated code
compiles against the provider's pinned SDK without a local SDK override.
The probe validates the framework schema and executes Create, Read, Update,
single-UUID import and Archive against a local HTTP server. Create returns a
UUID; every subsequent lifecycle HTTP request must address that exact UUID.
No `feature_flag_id` configuration or credentials are supplied. A negative
control changes Archive's argument to a different UUID and requires the HTTP path
assertion to fail, so an unrelated compilation failure cannot satisfy it.

This is maintained local/mock evidence, derived from the feature flag prototype;
it establishes no live-API behavior. The separate feature flag sandbox with its
local SDK extension exercises the proposed Terraform-specific Update operation.
