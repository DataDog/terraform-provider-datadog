---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291307299"
  confluence_id: "7291830358"
---

# Generating a data source

**Who this is for:** you want practitioners to read an object that already exists:
created elsewhere, or by a resource.

> **Status check before you invest.** tfgen generates data sources and the golden
> snapshots cover them, but **none have been adopted into the provider**.
> `generatedDatasources` is an empty slice. You would be the first. The generator
> path is real and tested; the social path is not worn in. See
> [reference/scope.md](../reference/scope.md).

## Singular or plural

```yaml
tags:
  - name: Actions Datastores
    x-datadog-tf-generator:
      datastore:
        datasource_singular:
          group:
            read: GetDatastore
            search: ListDatastores
        datasource_list:
          group:
            read: ListDatastores
```

A singular data source can declare read, search, or both. For a list data source,
put the list endpoint in `group.read`. The family name supplies the singular
artifact name; the list block pluralizes it unless `artifact_name` overrides it.

## A singular search must match exactly one

Zero matches is an error. Two matches is an error. That is deliberate: a data
source that silently picked the first of several would produce configurations that
change meaning as data changes. So choose a filter practitioners can make unique,
and say so in `tf_description`.

If the only available filter cannot be unique, you want a plural data source.

## Combining read and search

When both are present the generated data source uses the ID when configured and the
filters otherwise. The two response models must be compatible. A by-ID response and
a list item that disagree about the object's shape cannot project into one Terraform
schema, and the merge will say so.

## Identity

Data-source identity is `data.id` only. The annotation accepts other `id_strategy`
values and emission rejects them; see [reference/scope.md](../reference/scope.md).

## Writing the test is the real work

A data source reads something that already exists, so the test has to arrange for
that something. Prefer creating it in the same Terraform configuration, ideally
with the corresponding resource, rather than depending on permanent objects in the
test organisation. A data source whose matching resource is also generated is the
easiest thing to test, because the test can create exactly what it reads.

Continue at [shipping/](../shipping/README.md).
