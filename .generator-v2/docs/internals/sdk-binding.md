---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
  confluence_id: "7291503766"
---

# Stage 3 — Bind the Go SDK call

**Who this is for:** you are changing how generated code reaches the Datadog Go SDK.

Packages: `internal/sdkbind/` and `internal/sdkbinding/`.

## Two representations

- **`SDKOperationBinding`** hangs off an `Operation`. Derived from OpenAPI metadata:
  required and optional arguments, Go types, call order, optional-parameter setters.
- **`SDKCall`** hangs off an artifact lifecycle role. Emitter-ready: package, API
  struct, method, request and response types, arguments, result item type, and
  pagination behaviour.

## OpenAPI is authoritative

This is the rule to understand before changing anything here.

The binding is derived from the OpenAPI operation. The provider's pinned Go SDK is
loaded as **best-effort corroboration only**. A missing package or a newly added
endpoint does not stop generation; a disagreement produces a warning and the
OpenAPI-derived binding wins.

The consequence is intentional: generation can succeed for an endpoint the pinned SDK
does not have yet, and the failure surfaces at `make build` instead. SDK update and
provider build are separate merge gates. A valid annotation is therefore not
sufficient on its own — the endpoint must also exist in the SDK the provider pins.

Corroboration is also why some generator tests are `integration`-tagged: they shell
out and need the provider module's dependency graph on disk. `make
tfgen-test-integration` runs `go mod download` first for exactly that reason.

## API instance accessors

Generated `Configure` prefers an existing accessor from the provider's `ApiInstances`
helper, found relative to the output root at
`<output-root>/../internal/utils/api_instances_helper.go`. When one exists you get
`providerData.DatadogApiInstances.GetIncidentsApiV2()`.

When it is absent, the run warns and falls back to constructing the API from the
HTTP client: `datadogV2.NewIncidentsApi(providerData.DatadogApiInstances.HttpClient)`.

This bites fixture harnesses: generating into a bare temp directory takes the
fallback and produces output that differs from a golden recorded against the
provider checkout. `internal/providermod` answers where the provider module is.
