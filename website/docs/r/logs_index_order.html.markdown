---
layout: "datadog"
page_title: "Datadog: datadog_logs_index_order"
sidebar_current: "docs-datadog-resource-logs-index-order"
description: |-
  Provides a Datadog logs index order resource. This can be used to manage the order of logs indexes.
---

# datadog_logs_index_order

Provides a Datadog [Logs Index API](https://docs.datadoghq.com/api/?lang=python#logs-indexes) resource. This can be used to manage the order of Datadog logs indexes.

## Example Usage

```hcl
resource "datadog_logs_index_order" "sample_index_order" {
    name = "sample_index_order"
    depends_on = [
        "datadog_logs_index.sample_index"
    ]
    indexes = [
        "${datadog_logs_index.sample_index.id}"
    ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The unique name of the index order resource. 
* `indexes` - (Required) The index resource list. Logs are tested against the query filter of each index one by one following the order of the list.

## Import

The current Datadog Terraform provider version does not support the creation and deletion of index orders. 
Do `terraform import <datadog_logs_index_order.name> <name>` to import index order to Terraform. There must be at
most one `datadog_logs_index_order` resource.
