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

## Schema

### Required

- **name** (String, Required) Your archive name.
- **query** (String, Required) The archive query/filter. Logs matching this query are included in the archive.

### Optional

- **azure** (Map of String, Optional, Deprecated) Definition of an azure archive.
- **azure_archive** (Block List, Max: 1) Definition of an azure archive. (see [below for nested schema](#nestedblock--azure_archive))
- **gcs** (Map of String, Optional, Deprecated) Definition of a GCS archive.
- **gcs_archive** (Block List, Max: 1) Definition of a GCS archive. (see [below for nested schema](#nestedblock--gcs_archive))
- **id** (String, Optional) The ID of this resource.
- **include_tags** (Boolean, Optional) To store the tags in the archive, set the value `true`. If it is set to `false`, the tags will be dropped when the logs are sent to the archive.
- **rehydration_tags** (List of String, Optional) An array of tags to add to rehydrated logs from an archive.
- **s3** (Map of String, Optional, Deprecated) Definition of an s3 archive.
- **s3_archive** (Block List, Max: 1) Definition of an s3 archive. (see [below for nested schema](#nestedblock--s3_archive))

<a id="nestedblock--azure_archive"></a>
### Nested Schema for `azure_archive`

Required:

- **client_id** (String, Required) Your client id.
- **container** (String, Required) The container where the archive will be stored.
- **storage_account** (String, Required) The associated storage account.
- **tenant_id** (String, Required) Your tenant id.

Optional:

- **path** (String, Optional) The path where the archive will be stored.


<a id="nestedblock--gcs_archive"></a>
### Nested Schema for `gcs_archive`

Required:

- **bucket** (String, Required) Name of your GCS bucket.
- **client_email** (String, Required) Your client email.
- **path** (String, Required) Path where the archive will be stored.
- **project_id** (String, Required) Your project id.


<a id="nestedblock--s3_archive"></a>
### Nested Schema for `s3_archive`

Required:

- **account_id** (String, Required) Your AWS account id.
- **bucket** (String, Required) Name of your s3 bucket.
- **path** (String, Required) Path where the archive will be stored.
- **role_name** (String, Required) Your AWS role name

## Import

Import is supported using the following syntax:

```shell
terraform import datadog_logs_archive.my_s3_archive 1Aabc2_dfQPLnXy3HlfK4hi
```
