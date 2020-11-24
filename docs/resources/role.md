---
page_title: "datadog_role"
---

# datadog_role Resource

Provides a Datadog role resource. This can be used to create and manage Datadog roles.

## Example Usage

```hcl
# Source the permissions
data "datadog_permissions" "bar" {}

# Create a new Datadog role
resource "datadog_role" "foo" {
  name  = "foo"
  permission {
    id = data.datadog_permissions.bar.permissions.monitors_downtime
 }
  permission {
    id = data.datadog_permissions.bar.permissions.monitors_write
 }
}
```

## Argument Reference

The following arguments are supported:

-   `name`: (Required) The name of the role to create.
-   `permission`: (Optional) Blocks containing permission ID to grant to the role.

## Attributes Reference

The following attributes are exported:

-   `user_count`: The number of users that have this role.

## Import

Roles can be imported using their ID, e.g.

```
$ terraform import datadog_role.example_role 000000-0000-0000-0000-000000000000
```
