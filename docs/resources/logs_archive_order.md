---
page_title: "datadog_logs_archive_order"
---

# datadog_logs_archive_order Resource

Provides a Datadog [Logs Archive API](https://docs.datadoghq.com/api/v2/logs-archives/) resource, which is used to manage Datadog log archives order.

## Example Usage

```hcl
resource "datadog_logs_archive_order" "sample_archive_order" {
    name = "sample_archive_order"
    depends_on = [
        "datadog_logs_archive.sample_archive_1",
        "datadog_logs_archive.sample_archive_2"
    ]
    archives = [
        "${datadog_logs_archive.sample_archive_2.id}",
        "${datadog_logs_archive.sample_archive_1.id}"
    ]
}
```

## Argument Reference

The following arguments are supported:

- `name`: (Required) The name attribute in the resource `datadog_logs_archive_order` needs to be unique. It's recommended to use the same value as the resource `NAME`. No related field is available in [Logs Archive API](https://docs.datadoghq.com/api/v2/logs-archives/#get-archive-order).
- `archives`: (Required) The archive IDs list. The order of archive IDs in this attribute defines the overall archive order for logs.

## Attributes Reference

- `archives`: The `archives` list contains the IDs of resources created and imported by the [datadog_logs_archive](logs_archive.html#datadog_logs_archive). Updating the order of archives in this list reflects the application order of the archives. You cannot delete or create archive by deleting or adding IDs to this list.

## Import

There must be at most one `datadog_logs_archive_order` resource. Archive order creation is not supported from logs config API. You can import the `datadog_logs_archive_order` or create an archive order (which is actually doing the update operation).

```
terraform import <datadog_logs_archive_order.name> <name>
```
