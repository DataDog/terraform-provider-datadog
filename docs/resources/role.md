---
page_title: "datadog_role Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog role resource. This can be used to create and manage Datadog roles.
---

# Resource `datadog_role`

Provides a Datadog role resource. This can be used to create and manage Datadog roles.

## Example Usage

```terraform
# Source the permissions
data "datadog_permissions" "bar" {}

# Create a new Datadog role
resource "datadog_role" "foo" {
  name = "foo"
  permission {
    id = data.datadog_permissions.bar.permissions.monitors_downtime
  }
  permission {
    id = data.datadog_permissions.bar.permissions.monitors_write
  }
}
```

## Schema

### Required

- **name** (String, Required) Name of the role.

### Optional

- **id** (String, Optional) The ID of this resource.
- **permission** (Block Set) Set of objects containing the permission ID and the name of the permissions granted to this role. (see [below for nested schema](#nestedblock--permission))

### Read-only

- **user_count** (Number, Read-only) Number of users that have this role.

<a id="nestedblock--permission"></a>
### Nested Schema for `permission`

Required:

- **id** (String, Required) ID of the permission to assign.

Read-only:

- **name** (String, Read-only) Name of the permission.

## Import

Import is supported using the following syntax:

```shell
# Roles can be imported using their ID, e.g.
terraform import datadog_role.example_role 000000-0000-0000-0000-000000000000
```
