---
page_title: "datadog_role Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve information about an existing role for use in other resources.
---

# Data Source `datadog_role`

Use this data source to retrieve information about an existing role for use in other resources.

## Example Usage

```terraform
data "datadog_role" "test" {
  filter = "Datadog Standard Role"
}
```

## Schema

### Required

- **filter** (String, Required) A string on which to filter the roles.

### Optional

- **id** (String, Optional) The ID of this resource.

### Read-only

- **name** (String, Read-only) Name of the role.
- **user_count** (Number, Read-only) Number of users assigned to this role.


