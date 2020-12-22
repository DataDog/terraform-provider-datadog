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

- **name** (String, Required)
- **query** (String, Required)

### Optional

- **azure** (Map of String, Optional)
- **gcs** (Map of String, Optional)
- **id** (String, Optional) The ID of this resource.
- **include_tags** (Boolean, Optional)
- **rehydration_tags** (List of String, Optional)
- **s3** (Map of String, Optional)

## Import

Import is supported using the following syntax:

```shell
# Logs archives can be imported using their public string ID, e.g.

terraform import datadog_logs_archive.my_s3_archive 1Aabc2_dfQPLnXy3HlfK4hi
```
