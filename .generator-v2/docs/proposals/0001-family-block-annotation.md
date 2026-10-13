# Proposal 0001: Declare Terraform families on their API tag

- **Status:** implemented alongside legacy operation annotations.
- **Scope:** the `x-datadog-tf-generator` annotation surface. Generated artifact behavior is unchanged.

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

The document-root registry proposal is replaced by this design. Legacy
operation-level annotations and property-level `sensitive` annotations remain
supported. New family declarations use Tag Objects.

## Names and blocks

A family key is lowercase snake_case, at most 64 characters, without a `datadog_`
prefix. It is the default artifact name for both the resource and singular data
source; Terraform keeps these in separate namespaces. A list data source
pluralizes the final word (`datastore` → `datastores`, `logs_index` →
`logs_indexes`, `monitor_config_policy` → `monitor_config_policies`).

Each block can override `artifact_name`. Derived names must satisfy the same
constraints as explicit names. Global artifact-name checks detect collisions
across tags and between tag declarations and legacy operation annotations.

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
spec fragment. It collects group operation IDs from tag registries and legacy
annotations. It also compares tag registries directly during diffing so changes
to names, descriptions, groups, or skip flags dispatch Terraform even if no
operation changed.

## Migration and validation

Merge transformer and provider support first, in either order. Then move the
five spec families under their existing tags. The spec merge script needs no
Terraform-specific root-extension handling.

Tests cover naming, pluralization, skipping, malformed declarations, multiple
tags, duplicate families/tags, and legacy compatibility. The source-equivalence
test moves an incident-type resource declaration from its legacy GET annotation
onto the `Incidents` tag and compares generated Go source byte for byte. Its
synthetic anchor is POST, proving the output does not depend on the declaration's
original position.

The migration retains all five families and 17 operation IDs: four integration
resources are enabled and the datastore list data source remains skipped.
