---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
  confluence_id: "7291307299"
---

# Generating a Terraform artifact

**Who this is for:** you own a Datadog v2 API endpoint and want it available in the
Terraform provider without writing the provider code yourself.

You declare a Terraform family on an API tag in `datadog-api-spec`. Generation produces the provider
code, an example, and an acceptance-test scaffold. A provider maintainer, possibly
you, then finishes the test and merges it.

Before annotating, confirm your endpoint is in scope: [reference/scope.md](../reference/scope.md).

## Pick the artifact kind first

| A practitioner should be able to… | Kind | Needs |
|---|---|---|
| Create, update, and destroy the object through Terraform | **Resource** | create + read endpoints at minimum; update and delete if they exist |
| Look up an object that already exists, by ID or filter | **Data source**, singular | a by-ID read, a list endpoint, or both |
| Enumerate objects | **Data source**, plural | a list endpoint |

Resources are the shipping path: see [resources.md](resources.md).
Data sources are: [data-sources.md](data-sources.md)

A resource and a data source may share a name: they are separate Terraform
namespaces. `artifact_name` must be unique *within* its kind.

## The annotation, minimally

Declare the family once on its owning API tag. Each group value is an exact
`operationId`; the transformer uses those references to dispatch Terraform.

```yaml
tags:
  - name: Incidents
    x-datadog-tf-generator:
      incident_type:
        resource:
          tf_description: Provides a Datadog incident type resource.
          group:
            create: CreateIncidentType
            read: GetIncidentType
            update: UpdateIncidentType
            delete: DeleteIncidentType
```

Use `resource`, `datasource_singular`, or `datasource_list` blocks. A family may
contain all three. Resources require create, read and delete; update is optional.
Singular data sources require read or search; list data sources require read.
See [the annotation reference](../reference/annotation.md) for all accepted fields.

## Replacing something hand-written

If the provider already ships a hand-written artifact under your name, generation
will refuse to overwrite it. A file with no tfgen marker is never silently
replaced. Authorise the takeover explicitly:

```yaml
overwrites: NewIncidentTypeResource
```

That both permits the overwrite and rewires provider registration. Omit it for
purely additive generation. Be aware a takeover can touch more than one file: if the
hand-written resource declares model types its sibling data source uses, replacing
one file alone breaks the package.

## Opting out without removing the annotation

```yaml
skip: true
```

Keeps the annotation in the spec and out of generation, so the choice is visible to
reviewers rather than inferred from an absence.

## Marking a field sensitive

Use `x-secret: true` on a Schema Object for display redaction, or `writeOnly: true`
for write-only handling. Both mark the field sensitive. Legacy
`x-datadog-tf-generator.sensitive` overrides are no longer read. See
[secret handling](../reference/scope.md#secrets-and-sensitivity) for state behavior.

## Check before you open the pull request

- The artifact name is lowercase `snake_case` and does not include `datadog_`.
- Every `operationId` under `group` exists and is spelled exactly.
- The endpoint and its models are in the Go SDK version the provider pins.
- Response properties have useful `description`s, they become the Terraform schema
  descriptions practitioners read.

## What happens next

Generation runs from the API-spec pull request, before it merges, and produces
provider changes. For what triggers it and who owns that machinery, see
[pipeline.md](../pipeline.md).

The result is a **draft** provider pull request that is not yet correct-by-proof.
Someone has to finish it: [shipping/](../shipping/README.md).
