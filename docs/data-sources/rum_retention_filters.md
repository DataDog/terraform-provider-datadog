---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_rum_retention_filters Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog RUM retention filters datasource. This can be used to retrieve all RUM retention filters for a given RUM application.
---

# datadog_rum_retention_filters (Data Source)

Provides a Datadog RUM retention filters datasource. This can be used to retrieve all RUM retention filters for a given RUM application.

## Example Usage

```terraform
data "datadog_rum_retention_filters" "testing_rum_retention_filters" {
  application_id = "<APPLICATION_ID>"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `application_id` (String) RUM application ID.

### Read-Only

- `id` (String) The ID of this resource.
- `retention_filters` (List of Object) The list of RUM retention filters. (see [below for nested schema](#nestedatt--retention_filters))

<a id="nestedatt--retention_filters"></a>
### Nested Schema for `retention_filters`

Read-Only:

- `enabled` (Boolean)
- `event_type` (String)
- `id` (String)
- `name` (String)
- `query` (String)
- `sample_rate` (Number)
