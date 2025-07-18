---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_security_monitoring_filter Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Security Monitoring Rule API resource for security filters.
---

# datadog_security_monitoring_filter (Resource)

Provides a Datadog Security Monitoring Rule API resource for security filters.

## Example Usage

```terraform
resource "datadog_security_monitoring_filter" "my_filter" {
  name = "My filter"

  query      = "The filter is filtering."
  is_enabled = true

  exclusion_filter {
    name  = "first"
    query = "exclude some logs"
  }

  exclusion_filter {
    name  = "second"
    query = "exclude some other logs"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `is_enabled` (Boolean) Whether the security filter is enabled.
- `name` (String) The name of the security filter.
- `query` (String) The query of the security filter.

### Optional

- `exclusion_filter` (Block List) Exclusion filters to exclude some logs from the security filter. (see [below for nested schema](#nestedblock--exclusion_filter))
- `filtered_data_type` (String) The filtered data type. Valid values are `logs`. Defaults to `"logs"`.

### Read-Only

- `id` (String) The ID of this resource.
- `version` (Number) The version of the security filter.

<a id="nestedblock--exclusion_filter"></a>
### Nested Schema for `exclusion_filter`

Required:

- `name` (String) Exclusion filter name.
- `query` (String) Exclusion filter query. Logs that match this query are excluded from the security filter.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# Security monitoring filters can be imported using ID, e.g.
terraform import datadog_security_monitoring_filter.my_filter m0o-hto-lkb
```
