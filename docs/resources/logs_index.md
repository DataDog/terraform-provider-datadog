---
page_title: "datadog_logs_index Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Index API resource. This can be used to create and manage Datadog logs indexes.
---

# Resource `datadog_logs_index`

Provides a Datadog Logs Index API resource. This can be used to create and manage Datadog logs indexes.

## Example Usage

```terraform
# A sample Datadog logs index resource definition. Note that at this point, it is not possible to create new logs indexes through Terraform, so the name field must match a name of an already existing index. If you want to keep the current state of the index, we suggest importing it (see below).

resource "datadog_logs_index" "sample_index" {
  name = "your index"
  filter {
    query = "*"
  }
  exclusion_filter {
    name       = "Filter coredns logs"
    is_enabled = true
    filter {
      query       = "app:coredns"
      sample_rate = 0.97
    }
  }
  exclusion_filter {
    name       = "Kubernetes apiserver"
    is_enabled = true
    filter {
      query       = "service:kube_apiserver"
      sample_rate = 1.0
    }
  }
}
```

## Schema

### Required

- **filter** (Block List, Min: 1) Logs filter (see [below for nested schema](#nestedblock--filter))
- **name** (String, Required) The name of the index.

### Optional

- **exclusion_filter** (Block List) List of exclusion filters. (see [below for nested schema](#nestedblock--exclusion_filter))
- **id** (String, Optional) The ID of this resource.

<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

Required:

- **query** (String, Required) Logs filter criteria. Only logs matching this filter criteria are considered for this index.


<a id="nestedblock--exclusion_filter"></a>
### Nested Schema for `exclusion_filter`

Optional:

- **filter** (Block List) (see [below for nested schema](#nestedblock--exclusion_filter--filter))
- **is_enabled** (Boolean, Optional) A boolean stating if the exclusion is active or not.
- **name** (String, Optional) The name of the exclusion filter.

<a id="nestedblock--exclusion_filter--filter"></a>
### Nested Schema for `exclusion_filter.filter`

Optional:

- **query** (String, Optional) Only logs matching the filter criteria and the query of the parent index will be considered for this exclusion filter.
- **sample_rate** (Number, Optional) The fraction of logs excluded by the exclusion filter, when active.

## Import

Import is supported using the following syntax:

```shell
terraform import <datadog_logs_index.name> <indexName>
```
