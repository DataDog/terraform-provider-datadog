---
page_title: "datadog_logs_archive Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Archive API resource, which is used to create and manage Datadog logs archives.
---

# Resource `datadog_logs_archive`

Provides a Datadog Logs Archive API resource, which is used to create and manage Datadog logs archives.

## Example Usage

```terraform
# Create a Datadog logs archive:

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

## Schema

### Required

- **name** (String) Your archive name.
- **query** (String) The archive query/filter. Logs matching this query are included in the archive.

### Optional

- **azure** (Map of String) Definition of an azure archive.
- **gcs** (Map of String) Definition of a GCS archive.
- **id** (String) The ID of this resource.
- **include_tags** (Boolean) To store the tags in the archive, set the value `true`. If it is set to `false`, the tags will be dropped when the logs are sent to the archive.
- **rehydration_tags** (List of String) An array of tags to add to rehydrated logs from an archive.
- **s3** (Map of String) Definition of an s3 archive.

## Import

Import is supported using the following syntax:

```shell
# Logs archives can be imported using their public string ID, e.g.

terraform import datadog_logs_archive.my_s3_archive 1Aabc2_dfQPLnXy3HlfK4hi
```
