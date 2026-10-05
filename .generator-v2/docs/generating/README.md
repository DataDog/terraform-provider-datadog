---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
  confluence_id: "7291307299"
---

# Generating a Terraform artifact

**Who this is for:** you own a Datadog v2 API endpoint and want it available in the
Terraform provider without writing the provider code yourself.

You annotate an operation in `datadog-api-spec`. Generation produces the provider
code, an example, and an acceptance-test scaffold. A provider maintainer — possibly
you — then finishes the test and merges it.

Before annotating, confirm your endpoint is in scope: [reference/scope.md](../reference/scope.md).

## Pick the artifact kind first

This is the decision everything else follows from, and it is not about how many
endpoints you have — it is about what a practitioner does with the result.

| A practitioner should be able to… | Kind | Needs |
|---|---|---|
| Create, update, and destroy the object through Terraform | **Resource** | create + read endpoints at minimum; update and delete if they exist |
| Look up an object that already exists, by ID or filter | **Data source**, singular | a by-ID read, a list endpoint, or both |
| Enumerate objects | **Data source**, plural | a list endpoint |

Resources are the shipping path — see [resources.md](resources.md). Data sources are
generated but none have been adopted yet; [data-sources.md](data-sources.md) says
what that means for you.

A resource and a data source may share a name: they are separate Terraform
namespaces. `artifact_name` must be unique *within* its kind.

## The annotation, minimally

Two extensions go on the operation. `x-datadog-generators` routes the spec change to
the Terraform generator; `x-datadog-tf-generator` describes the artifact.

```yaml
paths:
  /api/v2/incidents/config/types:
    post:
      operationId: CreateIncidentType

      x-datadog-generators:
        generators:
          terraform: true

      x-datadog-tf-generator:
        artifact_kind: resource
        artifact_name: incident_type
        tf_description: Provides a Datadog incident type resource.
        group:
          create: CreateIncidentType
          read:   GetIncidentType
          update: UpdateIncidentType
          delete: DeleteIncidentType
```

Every field is documented, with its exact validation rules, in
[reference/annotation.md](../reference/annotation.md) — generated from the schema the
generator validates against, so it cannot describe a field that does not exist.

### Annotate the primary operation only

One operation owns the artifact and carries both extensions. Operations referenced
through `group` do not repeat them — they are named by `operationId` and found.

- Resource → annotate the **create** operation.
- Singular data source, by-ID → annotate the **by-ID read**.
- Singular data source, filter lookup → annotate the **list** operation.
- Plural data source → annotate the **list** operation.

Every value under `group` is an exact `operationId`. Not a path, not a method.

## Replacing something hand-written

If the provider already ships a hand-written artifact under your name, generation
will refuse to overwrite it — a file with no tfgen marker is never silently
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

`sensitive` goes on the **Schema Object**, not the operation annotation:

```yaml
some_secret:
  type: string
  x-datadog-tf-generator:
    sensitive: true
```

Usually you do not need it: `writeOnly: true` and `x-secret: true` already default a
field to sensitive. Read the secrets section of
[reference/scope.md](../reference/scope.md) before relying on either — redaction and
keeping values out of state are different things.

## Check before you open the pull request

- The artifact name is lowercase `snake_case` and does not include `datadog_`.
- Every `operationId` under `group` exists and is spelled exactly.
- The endpoint and its models are in the Go SDK version the provider pins.
- Response properties have useful `description`s — they become the Terraform schema
  descriptions practitioners read.
- You have not used `ignore`. It does not exist, and it aborts the entire run —
  including other teams' artifacts. See [reference/scope.md](../reference/scope.md).

## What happens next

Generation runs from the API-spec pull request, before it merges, and produces
provider changes. For what triggers it and who owns that machinery, see
[pipeline.md](../pipeline.md).

The result is a **draft** provider pull request that is not yet correct-by-proof.
Someone has to finish it: [shipping/](../shipping/README.md).
