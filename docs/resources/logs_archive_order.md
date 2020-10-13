---
page_title: "datadog_logs_archive_order"
---

# datadog_logs_archive_order Resource

Provides a Datadog [Logs Archive API](https://docs.datadoghq.com/api/v2/logs-archives/) resource, which is used to manage Datadog log archives order.

## Example Usage

```hcl
resource "datadog_logs_archive_order" "sample_archive_order" {
    archive_ids = [
        "${datadog_logs_archive.sample_archive_1.id}",
        "${datadog_logs_archive.sample_archive_2.id}"
    ]
}
```

## Argument Reference

The following arguments are supported:

- `archive_ids`: (Optional, Computed) The archive IDs list. The order of archive IDs in this attribute defines the overall archive order for logs. If `archive_ids` is empty or not specified, it will import the actual archive order, and create the resource. Otherwise, it will try to update the order.

## Attributes Reference

- `archive_ids`: The `archive_ids` list contains the IDs of resources created and imported by the [datadog_logs_archive](logs_archive.html#datadog_logs_archive). Updating the order of archives in this list reflects the application order of the archives. You cannot delete or create archive by deleting or adding IDs to this list.

## Import

There must be at most one `datadog_logs_archive_order` resource. You can import the `datadog_logs_archive_order` or create an archive order.

```
terraform import <datadog_logs_archive_order.name> archiveOrderID
```
