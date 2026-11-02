# 0005: Separate the caller-assigned create name from API identity

For generator maintainers, this record defines identity for APIs requiring a caller-assigned JSON:API `data.id`.

- **Date:** 2026-11-02
- **Status:** proposed in #4460; no production resource has adopted this contract.

## Decision

Expose required `name` with replacement-on-change semantics and retain computed
`id`. Create sends `name`; response mapping stores the returned `id`. Subsequent
read, update, and delete calls use `id`. Update reads that ID from prior state
because the computed value in its plan may be unknown. Update body IDs also use
the API ID, so the body and path agree.

This follows the separate name/ID shape of the hand-written metric resources.
Unlike blindly copying their response mapping, refresh preserves the configured
name: overwriting a required input with a normalized value violates Terraform's
plan/state consistency contract. Keeping a single required `id` instead would
lose canonical identity and assume the API accepts the original input forever.

## Naming, import, and compatibility

`name` is the deterministic input name for all such resources. If an attribute
or parent path parameter already claims it, fail generation with a diagnostic;
do not silently alias unrelated fields. A configurable mapping can be designed
later if an actual API needs it. Require a response ID to recover API identity.

Import takes the API ID, as other resources do. The first refresh fills a null
name from that ID, since the original spelling cannot be recovered. Users adopt
that name in configuration. Existing configured names are never rewritten.

No production resource is registered by this change. Changing the unshipped
fixture now avoids a later breaking rename of a configurable `id`.

## Validation

The offline lifecycle fixture returns an unchanged or canonicalized ID and only
serves subsequent operations at the returned ID. It verifies PUT body identity,
unknown computed identity during update, configured-name preservation, import,
and replacement when the name changes. Enum identity must retain its named SDK
type and use the validating constructor; unsupported enum shapes fail closed.
Live acceptance remains a prerequisite for adopting a production resource and
must establish that API's normalization and import behavior.
