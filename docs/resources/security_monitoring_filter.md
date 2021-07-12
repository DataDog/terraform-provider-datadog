---
page_title: "datadog_security_monitoring_filter Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Security Monitoring Filter API resource. This can be used to create and manage Datadog security monitoring filters

# Resource `datadog_security_monitoring_filter`

Provides a Datadog Security Monitoring Filter API resource. This can be used to create and manage Datadog security monitoring filters.

## Example Usage

```terraform
resource "datadog_security_monitoring_filter" "myfilter" {
  name = "My filter"

  query = "The filter is filtering."
  is_enabled = true

  exclusion_filter {
    name = "first"
    query = "exclude some logs"
  }

  exclusion_filter {
    name = "second"
    query = "exclude some other logs"
  }
  
}
```

## Schema

### Required

- **is_enabled** (String, Required) The name of the security filter.
- **query** (String, Required) The query of the security filter.
- **is_enabled** (Boolean, Required) Whether the security filter is enabled.

### Optional

- **exclusion_filter** (Block List, Optional) Exclusion filters to exclude some logs from the security filter. (see [below for nested schema](#nestedblock--filter))
- **filtered_data_type** (String, Optional) The filtered data type (Default to 'logs').
  
<a id="nestedblock--filter"></a>
### Nested Schema for `exclusion_filter`

Required:

- **query** (String, Required) Exclusion filter query. Logs that match this query are excluded from the security filter.
- **name** (String, Required) Exclusion filter name.

## Import

Import is supported using the following syntax:

```shell
# Security monitoring rules can be imported using ID, e.g.
terraform import datadog_security_monitoring_filter.my_filter m0o-hto-lkb
```
