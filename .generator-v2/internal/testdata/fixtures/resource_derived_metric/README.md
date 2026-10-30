# Required caller-assigned identity

Reduced derived-metric CRUD spec from the TF-426 assessment. The authoritative
create schema requires `data.id`; no production resource is registered by this fixture.

The golden test checks generated output. The integration test generates twice,
checks stability, then runs `identity_test.go.txt` against the generated resource
with a local HTTP server. It covers required ID configuration, request identity,
replacement modifiers, CRUD state, and import/refresh without live API calls.
