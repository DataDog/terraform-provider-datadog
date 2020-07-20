---
layout: "datadog"
page_title: "Datadog: datadog_logs_archive"
sidebar_current: "docs-datadog-resource-logs-archive"
description: |-
  Provides a Datadog logs archive resource, which is used to create and manage logs archives.
---

# datadog_logs_archive

Provides a Datadog [Logs Archive API](https://docs.datadoghq.com/api/v2/logs-archives/) resource, which is used to create and manage Datadog logs archives.


## Example Usage

Create a Datadog logs archive:

```hcl
resource "datadog_logs_archive" "my_s3_archive" {
  name  = "my s3 archive"
  query = "service:myservice"
  s3 = {
    bucket     = "my-bucket"
    path       = "/path/foo"
    account_id = "001234567888"
    role_name  = "my-role-name"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Your archive name.
* `query` - (Required) The archive query/filter. Logs matching this query are included in the archive.
* `s3` - (Optional) Definition of an s3 archive.
  * `bucket` - (Required) Name of your s3 bucket.
  * `path` - (Optional, default = "") Path where the archive will be stored.
  * `account_id` - (Required) Your AWS account id.
  * `role_name` - (Required) Your AWS role name.
* `gcs` - (Optional) Definition of an gcs archive.
  * `bucket` - (Required) Name of your gcs bucket.
  * `path` - (Optional, default = "") Path where the archive will be stored.
  * `client_email` - (Required) Your client email.
  * `project_id` - (Required) Your project id.
* `azure` - (Optional) Definition of an azure archive.
  * `container` - (Required) The container where the archive will be stored.
  * `path` - (Optional, default = "") The path where the archive will be stored.
  * `tenant_id` - (Required) Your tenant id.
  * `client_id` - (Required) Your client id.
  * `storage_account` - (Required) The associated storage account.


An archive definition must have one (and only one) of the three possible types defined: s3, gcs, azure.

## Import

Logs archives can be imported using their public string ID, e.g.

```
$ terraform import datadog_logs_archive.my_s3_archive 1Aabc2_dfQPLnXy3HlfK4hi
```
