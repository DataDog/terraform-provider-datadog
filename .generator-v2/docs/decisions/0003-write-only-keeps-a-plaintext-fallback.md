---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291274458"
---

# 0003: Write-only secrets keep a plaintext fallback

- **Date:** 2026-10-08
- **Status:** accepted, unmerged at the time of writing

## Context

Terraform's write-only attributes are a 1.11 feature. A schema that declares one
is rejected outright by earlier Terraform, with no partial degradation: the
practitioner cannot opt out of the attribute, because the provider's schema is
not negotiable per-configuration.

Generated resources exposed a write-only secret as exactly two attributes,
`<attr>_wo` and `<attr>_wo_version`, and deliberately omitted the stateful
plaintext `<attr>`. The reasoning was that a secret in Terraform state is the
problem write-only exists to solve, and that a generated resource — unlike the
hand-written ones, which predate write-only and carry plaintext for backwards
compatibility — has no existing users to keep compatible, so it should start
clean. [scope.md](../reference/scope.md) said so in as many words: *"Hand-written
resources use the legacy three-attribute pattern via `datadog/internal/fwutils`.
Do not copy that shape into generated resources."* The original specification,
[APIR-3279](https://datadoghq.atlassian.net/browse/APIR-3279), was explicit on
the same point: *"Do not generate a plaintext alias, fallback path,
private-state hash, digest comparison, or keepers mechanism"*, with "no original
plaintext Terraform field is present" as an acceptance criterion.

The cost of that is the whole resource, not the one attribute. Any resource with
a single write-only secret anywhere in its tree becomes unusable below Terraform
1.11 — including for practitioners who never configure that secret, and
including resources whose secret is optional. Datadog does not control its
users' Terraform version, so "start clean" is in practice "do not ship to a
share of the user base we cannot measure."

## Decision

Generated resources emit three attributes, not two: `<attr>_wo` and
`<attr>_wo_version` for Terraform 1.11+, plus the stateful plaintext `<attr>` as
the pre-1.11 fallback. `PreferWriteOnlyAttribute` on the plaintext attribute
steers a 1.11+ practitioner to the write-only path, so the fallback is reachable
without being the recommendation.

This is the shape hand-written resources already use, so `fwutils` needed a
generalization rather than a new mechanism. `WriteOnlySecretModeDual` is
`WriteOnlySecretModeLegacy` made to honour `Required`:

- a secret the API requires on create gets `ExactlyOneOf(<attr>, <attr>_wo)`
- an optional one gets `ConflictsWith`

Requiredness lands on the pair because either half alone satisfies the API;
marking `<attr>_wo` `Required` would reject a configuration that legitimately
sets only the plaintext attribute. `ModeLegacy` — which predates `Required` and
always forces `ExactlyOneOf` — and the write-only-only mode are both left
alone, so no already-shipped resource's schema moves.

## Consequences

- **The security posture of a generated resource is now the practitioner's
  choice, not the generator's.** A 1.11+ user who follows the schema's own
  guidance keeps the secret out of state; a user who sets `<attr>` puts it in
  state. That is the same bargain every hand-written Datadog resource already
  offers, which is the argument for it: the generator was holding generated
  resources to a stricter standard than the provider it ships into.
- **The three attributes are unconditional, by decision.** Every generated
  write-only secret gets all three, with no annotation or flag to request the
  write-only-only shape. This is deliberate and not a gap awaiting a follow-up:
  a per-secret opt-out would mean a generated resource's Terraform floor varied
  with the spec, so a practitioner could not tell from the provider version
  whether a given resource was usable.
- The four resources generated before this change — the Databricks, Elastic
  Cloud, Snowflake, and Twilio integration accounts — keep the two-attribute
  shape until they are regenerated, so they stay on Terraform 1.11+ in the
  meantime. Regenerating them adds a plaintext attribute to released
  resources, which is a user-visible change and belongs in its own record.
- [scope.md](../reference/scope.md) describes the three-attribute contract and
  the requiredness rule, and no longer tells a reader the two shapes must stay
  apart.

## Why this is recorded

The two-attribute shape was not an oversight; it was specified deliberately in
APIR-3279 and written down as a rule, in the imperative, in the one file that
states what the generator supports. That ticket is closed, so nothing about it
will now change to reflect this; this record is the only thing connecting its
requirement to the code that stopped honouring it. Anyone
reading only the new `scope.md` would find the opposite rule and no trace that
the old one was ever argued for — and the most likely next question about this
code is "why does a generated resource put a secret in state at all?", which the
reference page is the wrong place to answer.
