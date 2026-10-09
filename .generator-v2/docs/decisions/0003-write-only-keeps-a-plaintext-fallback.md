---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291274458"
---

# 0003: Write-only secrets keep a plaintext fallback

- **Date:** 2026-10-08
- **Status:** accepted, unmerged at the time of writing

## Context

For generator maintainers, this record explains why generated write-only secrets
also expose a stateful plaintext attribute.

[#4232](https://github.com/DataDog/terraform-provider-datadog/pull/4232)
introduced the two-attribute design: `<attr>_wo` and `<attr>_wo_version`, with no
plaintext fallback, to keep secrets out of Terraform state.

Write-only values require Terraform 1.11+. On older Terraform, a required
write-only attribute makes the entire resource unusable: omitting it fails the
required-argument check, and setting it fails the framework's version check.
An optional write-only attribute leaves the rest of the resource usable, but
users cannot configure that secret. Supporting users on older Terraform requires
a plaintext path.

## Decision

Every generated write-only secret gets a plaintext fallback, using the same
three-attribute pattern as hand-written resources. There is no per-secret opt-out.
Users on Terraform 1.11+ are guided toward the write-only path.

The [secret handling contract](../reference/scope.md#secrets-and-sensitivity)
describes the attributes, validators, and requiredness rules.

## Consequences

- Users who choose the plaintext attribute store the secret in Terraform state.
  The write-only path keeps it out of state. Compatibility therefore comes with
  an explicit choice about secret storage.
- Applying the fallback unconditionally avoids introducing different Terraform
  version requirements through per-secret generator settings.
- The previously generated Databricks, Elastic Cloud, Snowflake, and Twilio
  integration accounts retain their two-attribute schemas until regenerated.
  Adding their plaintext fallback is a separate, user-visible change.
- Existing hand-written callers retain Legacy mode and its validation behavior.

## Why this is recorded

The plaintext fallback reverses the deliberate choice in #4232 to exclude secrets
from generated resource state. This record preserves the compatibility rationale
for that change; the reference page describes the resulting contract.
