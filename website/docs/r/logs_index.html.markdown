---
layout: "datadog"
page_title: "Datadog: datadog_logs_index"
sidebar_current: "docs-datadog-resource-logs-index"
description: |-
  Provides a Datadog logs index resource. This can be used to create and manage logs indexes.
---

# datadog_logs_index

Provides a Datadog [Logs Index API](https://docs.datadoghq.com/api/?lang=python#logs-indexes) resource. This can be used to create and manage Datadog logs indexes.

## Example Usage:
Create a Datadog logs index resource.

```hcl
# Update a Datadog logs index
resource "datadog_logs_index" "sample_index" {
    name = "your index"
    filter {
        query = "*"
    }
    exclusion_filter {
        name = "Filter coredns logs"
        is_enabled = true
        filter {
            query = "app:coredns"
            sample_rate = 0.97
        }
    }
    exclusion_filter {
        name = "Kubernetes apiserver"
        is_enabled = true
        filter {
            query = "service:kube_apiserver"
            sample_rate = 1.0
        }
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the index.
* `filter` - (Required) Logs filter.
  * `query` - (Required) Logs filter criteria. Only logs matching this filter criteria are considered for this index.
* `exclusion_filter` - (Required) List of exclusion filter.
  * `filter` - (Required)
      * `query` - (Required) Only logs matching the filter criteria and the query of the parent index will be considered for this exclusion filter.
      * `sample_rate` - (Optional, default = 0.0) the fraction of logs excluded by the exclusion filter, when active.
  * `name` - (Optional) The name of the exclusion filter.
  * `is_enabled` - (Optional, default = false) A boolean stating if the exclusion is active or not.

## Import

The current datadog terraform provider version does not support the creation and deletion of index. 
To manage the existing indexes, do `terraform import <datadog_logs_index.name> <indexName>` to import them to terraform.

## Important Notes

The order of indexes is maintained in the separated resource [datadog_logs_index_order](logs_index_order.html#datadog_logs_index_order). 
