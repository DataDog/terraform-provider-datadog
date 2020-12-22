---
page_title: "datadog_logs_index_order Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Index API resource. This can be used to manage the order of Datadog logs indexes.
---

# Resource `datadog_logs_index_order`

Provides a Datadog Logs Index API resource. This can be used to manage the order of Datadog logs indexes.

## Example Usage

```terraform
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

## Schema

### Required

- **indexes** (List of String, Required)
- **name** (String, Required)

### Optional

- **id** (String, Optional) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# The Datadog Terraform Provider does not support the creation and deletion of index orders. There must be at most one `datadog_logs_index_order` resource
terraform import <datadog_logs_index_order.name> <name>
```
