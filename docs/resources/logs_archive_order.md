---
page_title: "datadog_logs_archive_order Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Archive API resource, which is used to manage Datadog log archives order.
---

# Resource `datadog_logs_archive_order`

Provides a Datadog Logs Archive API resource, which is used to manage Datadog log archives order.

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

- **archive_ids** (List of String, Optional)
- **id** (String, Optional) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# There must be at most one datadog_logs_archive_order resource. You can import the datadog_logs_archive_order or create an archive order.
terraform import <datadog_logs_archive_order.name> archiveOrderID
```
