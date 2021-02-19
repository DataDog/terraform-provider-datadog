---
page_title: "datadog_logs_archive_order Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Archive API https://docs.datadoghq.com/api/v2/logs-archives/ resource, which is used to manage Datadog log archives order.
---

# Resource `datadog_logs_archive_order`

Provides a Datadog [Logs Archive API](https://docs.datadoghq.com/api/v2/logs-archives/) resource, which is used to manage Datadog log archives order.

## Example Usage

```terraform
resource "datadog_logs_archive_order" "sample_archive_order" {
  archive_ids = [
    "${datadog_logs_archive.sample_archive_1.id}",
    "${datadog_logs_archive.sample_archive_2.id}"
  ]
}
```

## Schema

### Optional

- **archive_ids** (List of String, Optional) The archive IDs list. The order of archive IDs in this attribute defines the overall archive order for logs. If `archive_ids` is empty or not specified, it will import the actual archive order, and create the resource. Otherwise, it will try to update the order.
- **id** (String, Optional) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# There must be at most one datadog_logs_archive_order resource. You can import the datadog_logs_archive_order or create an archive order.
terraform import <datadog_logs_archive_order.name> archiveOrderID
```
