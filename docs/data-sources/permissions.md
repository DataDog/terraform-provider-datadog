---
page_title: "datadog_permissions Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve the list of Datadog permissions by name and their corresponding ID, for use in the role resource.
---

# Data Source `datadog_permissions`

Use this data source to retrieve the list of Datadog permissions by name and their corresponding ID, for use in the role resource.

## Example Usage

```terraform
data "datadog_permissions" "permissions" {}
```

## Schema

### Optional

- **id** (String) The ID of this resource.

### Read-only

- **permissions** (Map of String) Map of permissions names to their corresponding ID.


