---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291274458"
---

# 0004: Declare Terraform families on their API tag

For generator maintainers, this record explains the ownership model and rollout
of the tag-scoped Terraform annotation contract.

- **Date:** 2026-10-20

- **Status:** accepted; provider implementation pending merge in #4376.
- **Scope:** the `x-datadog-tf-generator` annotation surface. Generated artifact behavior is unchanged.

## Context and decision

An operation can participate in several artifacts, and an artifact can use
several operations. Keeping the declaration on a single endpoint obscures that
relationship and requires authors to repeat kind and cardinality fields.

Use each API's existing top-level tag as the owner of a family registry. This
keeps declarations with the owning team while allowing explicit references to
operations anywhere in the document. Support one declaration format; do not
introduce a dual-format compatibility period. The generated
[annotation reference](../reference/annotation.md) defines the accepted fields.

## Declaration

A tag owns a map of Terraform families. Each family can declare a resource, a
singular data source, and a list data source:

```yaml
tags:
  - name: Databricks Integration
    x-datadog-tf-generator:
      integration_databricks_account:
        resource:
          tf_description: Provides a Datadog Databricks integration account resource.
          group:
            create: CreateDatabricksIntegrationAccount
            read: GetDatabricksIntegrationAccount
            update: UpdateDatabricksIntegrationAccount
            delete: DeleteDatabricksIntegrationAccount
        datasource_singular:
          group:
            read: GetDatabricksIntegrationAccount
        datasource_list:
          group:
            read: ListDatabricksIntegrationAccounts
```

The extension belongs on the top-level Tag Object in `tags`, not on an operation's
list of tag names. Teams author it alongside their existing tag definition in
their own spec fragment. Ordinary tag export and concatenation preserve the
registry; no document-root extension merging is needed.

Family names are unique across the spec, including skipped families. A tag's
registry must be declared in one place. Multiple fragments may reference the
same tag, but cannot each contribute a registry for it. Conflicting declarations
are rejected rather than merged or silently overwritten.

The document-root registry proposal is replaced by this design. Operation-level
artifact declarations and property-level `sensitive` overrides are no longer
parsed. Sensitivity comes from `writeOnly` and `x-secret`. All families use Tag Objects.

## Names and blocks

A family key is lowercase snake_case, at most 64 characters, without a `datadog_`
prefix. It is the default artifact name for both the resource and singular data
source; Terraform keeps these in separate namespaces. A list data source
pluralizes the final word (`datastore` → `datastores`, `logs_index` →
`logs_indexes`, `monitor_config_policy` → `monitor_config_policies`).

Each block can override `artifact_name`. Derived names must satisfy the same
constraints as explicit names. Global artifact-name checks detect collisions
across all tag declarations.

| Block | Artifact kind | Group roles |
| --- | --- | --- |
| `resource` | Resource | `create`, `read`, `delete` required; `update` optional; no `search` |
| `datasource_singular` | Singular data source | `read` or `search` required; both allowed; no mutation roles |
| `datasource_list` | List data source | `read` required; no other roles |

Operation IDs remain explicit and resolve through the document-wide operation
index. Moving a declaration does not change the endpoints, SDK bindings, or
schemas used by generation. The tag organizes ownership; operations are not
inferred from its name.

Each block also supports `tf_description`, `id_strategy`, `overwrites`, and
`skip`. These fields retain their legacy meanings. A family's `skip: true`
disables all its blocks; a block's `skip: true` disables only that artifact.
Skipped declarations still undergo shape validation. A family must declare at
least one block or set `skip: true`; `skip: false` alone is not a declaration.

## Implementation

The parser reads `v3doc.Model.Tags[*].Extensions`, validates each registry against
[`family-registry.schema.json`](../../internal/contracts/family-registry.schema.json),
and converts each enabled block to existing `TrackingFieldMetadata`. Family
names and block kinds determine `artifact_kind`, `artifact_name`, and cardinality.
There is no separate generator path for tag declarations.

A synthetic operation carries each artifact's metadata. It inherits its primary
operation's path, method, tag, unstable marker, and raw schema handle:

- Resource: `group.create`.
- Singular data source: `group.read`, or `group.search` when no read is declared.
- List data source: `group.read`.

Synthetic operations are excluded from the operation-ID index so multiple
artifacts can share endpoints without changing group resolution. Tags and
families are processed deterministically; reordering tag declarations does not
change generation order. Diagnostics include the owning tag for invalid
registries and duplicate declarations.

The transformer selects the registry type by extension key and `Tag` parent,
validates its family/block fields, and retains source locations from each tag's
spec fragment. It collects group operation IDs from tag registries only. It also compares tag registries directly during diffing so changes
to names, descriptions, groups, or skip flags dispatch Terraform even if no
operation changed.

## Migration and validation

The rollout uses an annotation-free transition so the transformer can stop
accepting the old format before new declarations are introduced:

1. Remove the five legacy operation declarations from the spec and allow
   annotation-free build releases to accumulate. This landed in
   [datadog-api-spec#7066](https://github.com/ddoghq/datadog-api-spec/pull/7066).
2. Land transformer support and coordinate deployment of the matching validator.
   [openapi-transformer#516](https://github.com/DataDog/openapi-transformer/pull/516)
   is merged; merge status alone does not establish deployment status.
3. Merge provider support in
   [terraform-provider-datadog#4376](https://github.com/DataDog/terraform-provider-datadog/pull/4376).
4. Restore the declarations as tag families through
   [datadog-api-spec#7039](https://github.com/ddoghq/datadog-api-spec/pull/7039),
   once both consumers and the validator support the new format.

Historical build assets keep their old annotations. Rebase branches using those
merge bases onto an annotation-free or migrated baseline before generation.

The restoration preserves four enabled integration resources and the skipped
datastore list data source. Tests cover naming, pluralization, skipping,
malformed declarations, duplicate families/tags, metadata, and ignored legacy
annotations. Fixtures exercise the tag-scoped format. Missing primary operation
IDs fail their own artifact, appear in the report, and prevent orphan retirement
for that run.
