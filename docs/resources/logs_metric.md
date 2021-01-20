---
page_title: "datadog_logs_metric Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Resource for interacting with the logs_metric API
---

# Resource `datadog_logs_metric`

Resource for interacting with the logs_metric API

## Example Usage

```terraform
resource "datadog_logs_metric" "testing_logs_metric" {
  name = "testing.logs.metric"
  compute {
    aggregation_type = "distribution"
    path             = "@duration"
  }
  filter {
    query = "service:test"
  }
  group_by {
    path     = "@status"
    tag_name = "status"
  }
}
```

## Schema

### Required

- **compute** (Block List, Min: 1, Max: 1) The compute rule to compute the log-based metric. This field can't be updated after creation. (see [below for nested schema](#nestedblock--compute))
- **filter** (Block List, Min: 1, Max: 1) The log-based metric filter. Logs matching this filter will be aggregated in this metric. (see [below for nested schema](#nestedblock--filter))
- **name** (String, Required) The name of the log-based metric. This field can't be updated after creation.

### Optional

- **group_by** (Block List) The rules for the group by. (see [below for nested schema](#nestedblock--group_by))
- **id** (String, Optional) The ID of this resource.

<a id="nestedblock--compute"></a>
### Nested Schema for `compute`

Required:

- **aggregation_type** (String, Required) The type of aggregation to use. This field can't be updated after creation.

Optional:

- **path** (String, Optional) The path to the value the log-based metric will aggregate on (only used if the aggregation type is a "distribution"). This field can't be updated after creation.


<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

Required:

- **query** (String, Required) The search query - following the log search syntax.


<a id="nestedblock--group_by"></a>
### Nested Schema for `group_by`

Required:

- **path** (String, Required) The path to the value the log-based metric will be aggregated over.
- **tag_name** (String, Required) Name of the tag that gets created.

## Import

Import is supported using the following syntax:

```shell
terraform import datadog_logs_metric.testing_logs_metric testing.logs.metric
```
