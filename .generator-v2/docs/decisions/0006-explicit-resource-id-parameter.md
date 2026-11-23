# Explicit resource ID parameter: contract and rollout

Tracking: [TF-459](https://datadoghq.atlassian.net/browse/TF-459).

`id_strategy` selects ID extraction; resource-only `id_parameter` selects the
OpenAPI path argument receiving that ID in Read, Update, and Delete. The latter
is an author assertion of semantic compatibility with response `data.id`, not a
new extraction strategy. Binding occurs on artifact-local SDK calls before
schema and import construction. Omission preserves positional inference.

## Reproducible verification

- `make tfgen-test` covers contract/parser validation, lifecycle binding and
  diagnostics, generated output, artifact isolation, and unchanged fixture goldens.
- `make tfgen-test-integration` compiles generated resources, validates schemas,
  and runs maintained mock-API lifecycle coverage. The feature flag fixture and
  its runtime probe live in `internal/testdata/fixtures/resource_feature_flag/`.
- `make tfgen-update-goldens` regenerates fixture snapshots for review.
- `make tfgen-docs && make tfgen-docs-check` regenerates and checks the annotation
  reference from the contract.

The maintained fixture uses the provider's pinned SDK and a local HTTP server;
it needs neither the original prototype's SDK checkout nor live credentials.
It verifies that the UUID returned by Create reaches Get, Update, and Archive
without a separate configured feature flag ID. This is local/mock evidence.
Live feature flag acceptance remains a separate adoption check.

## Local sandbox evidence (2026-11-23)

The existing gitignored `sandbox/feature-flag-verification` was also regenerated
and checked with its local SDK override. Its manifest identifies source-spec
commit `f1f7380bbe4d74975ef9e92517fbd40215e74f78` and SDK base
`cdeaa1720dc83d695a363486d939dcd3a17b143e`. Input
`specs/tf459-feature-flag.json` retains the scoped feature flag operations and
reachable schemas, adding resource `id_parameter: feature_flag_id`.

| Input | SHA256 |
| --- | --- |
| Scoped annotated spec | `2ce8fc38f57c955c9a43ab4ed8a96bd28cbd95584ab1b9588a9510383d2333a5` |
| Local SDK `api/datadogV2/api_feature_flags.go` | `8e8a9e81a85e06968468af6624212d9e28a1df8d7a775066edfbcfa5b0ca43d3` |

From the provider root, using Go 1.26.10 and Terraform 1.13.1:

```sh
make tfgen-build
GOFLAGS=-modfile=$PWD/sandbox/feature-flag-verification/provider.mod \
  bin/tfgen generate \
  --spec sandbox/feature-flag-verification/specs/tf459-feature-flag.json \
  --output-root sandbox/feature-flag-verification/provider/datadog/fwprovider \
  --only-tags 'Feature Flags' \
  --report sandbox/feature-flag-verification/logs/tf459-generation-report.json
sandbox/feature-flag-verification/test-feature-flag.sh
sandbox/feature-flag-verification/build-provider.sh
```

`provider.mod` replaces `github.com/DataDog/datadog-api-client-go/v2` with the
sandbox's `datadog-api-client-go` directory. Generation reported one updated
artifact and zero failures. The sandbox's Make-based test script passed eight
tests, including direct HTTP lifecycle routing and a Terraform config without
`feature_flag_id` that imports a single UUID. Other preexisting sandbox probes
expect unrelated update/refresh defects; their passing results do not establish
that those defects were fixed.

The sandbox provider build passed. `terraform providers schema -json` from its
`schema-check` directory with `TF_CLI_CONFIG_FILE` pointing at `datadog.tfrc`
confirmed computed `id`, no required `id`, no `feature_flag_id`, and 19 top-level
attributes. Logs are under the sandbox's `logs/tf459-*`; generated Read, Update,
and Archive each parse `state.ID`, with passthrough ID import. These are local
artifacts and local/mock results, independent of the maintained fixture above.

Companion transformer changes are in
`src/openapi_transformer/openapi/x_datadog_tf_generator.py` and
`src/openapi_transformer/validation/rules/validate_x_datadog_tf_generator.py`.
The matching `tests/validate_spec/rules/test_validate_x_datadog_tf_generator.py`
suite passed 126 tests, including compilation/export preservation and resource-only
validation. The local companion checkout is based on
`0c20ae74ebefb95b973ae53a3c7d1b8b1ed973a0`; this is not a supporting release.

## Release order

1. Merge generator and OpenAPI transformer changes with passing consumer checks.
   Transformer modeling, validation, and compilation must accept and preserve
   the same resource-only annotation contract.
2. Release both components and record the actual supporting versions in TF-459
   or their PRs. Neither version is assigned by this implementation.
3. Only after both releases, adopt `id_parameter` in the authoritative feature
   flag source annotation and compile with the released transformer.
4. Regenerate the resource with the intended SDK version, review schema/import
   and Archive routing, and complete the live acceptance gate before shipping.

Source-spec adoption must wait for the two releases. Maintainer agreement,
release versions, and PR URLs are rollout evidence to record when available;
local implementation and mock tests do not establish them.
