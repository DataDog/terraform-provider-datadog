---
page_title: "datadog_logs_archive"
---

# datadog_logs_archive Resource

Provides a Datadog [Logs Archive API](https://docs.datadoghq.com/api/v2/logs-archives/) resource, which is used to create and manage Datadog logs archives.

## Example Usage

Create a Datadog logs archive:

```hcl
resource "datadog_logs_archive" "my_s3_archive" {
  name  = "my s3 archive"
  query = "service:myservice"
  s3_archive {
    bucket     = "my-bucket"
    path       = "/path/foo"
    account_id = "001234567888"
    role_name  = "my-role-name"
  }
}
```

## Argument Reference

The following arguments are supported:

-   `name`: (Required) Your archive name.
-   `query`: (Required) The archive query/filter. Logs matching this query are included in the archive.
-   `s3_archive`: (Optional) Definition of an s3 archive. List of one element with the structure below.
    -   `bucket`: (Required) Name of your s3 bucket.
    -   `path`: (Optional, default = "") Path where the archive will be stored.
    -   `account_id`: (Required) Your AWS account id.
    -   `role_name`: (Required) Your AWS role name.
-   `gcs_archive`: (Optional) Definition of an gcs archive. List of one element with the structure below.
    -   `bucket`: (Required) Name of your gcs bucket.
    -   `path`: (Optional, default = "") Path where the archive will be stored.
    -   `client_email`: (Required) Your client email.
    -   `project_id`: (Required) Your project id.
-   `azure_archive`: (Optional) Definition of an azure archive. List of one element with the structure below.
    -   `container`: (Required) The container where the archive will be stored.
    -   `path`: (Optional, default = "") The path where the archive will be stored.
    -   `tenant_id`: (Required) Your tenant id.
    -   `client_id`: (Required) Your client id.
    -   `storage_account`: (Required) The associated storage account.
-   `s3`: (Deprecated, Optional) Definition of an s3 archive. Use `s3_archive` instead.
-   `gcs`: (Deprecated, Optional) Definition of an gcs archive. Use `gcs_archive` instead.
-   `azure`: (Deprecated, Optional) Definition of an azure archive. Use `azure_archive` instead.
-   `rehydration_tags`: (Optional) An array of tags to add to rehydrated logs from an archive.
-   `include_tags`: (Optional, default=false) To store the tags in the archive, set the value "true". If it is set to "false", the tags will be dropped when the logs are sent to the archive.

An archive definition must have one (and only one) of the three possible types defined: s3, gcs, azure.

## Import

Logs archives can be imported using their public string ID, e.g.

```
$ terraform import datadog_logs_archive.my_s3_archive 1Aabc2_dfQPLnXy3HlfK4hi
```
