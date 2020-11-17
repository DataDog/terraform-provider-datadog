---
page_title: "datadog_role"
---

# datadog_role Resource

Provides a Datadog role resource. This can be used to create and manage Datadog roles.

## Example Usage

```hcl
# Create a new Datadog role
resource "datadog_role" "foo" {
  name  = "foo"
  permissions = [
    "${data.datadog_permissions.example_permission_1.id}",
    "${data.datadog_permissions.example_permission_2.id}"
]
}
```

## Argument Reference

The following arguments are supported:

-   `name`: (Required) The name of the role to create.
-   `permissions`: (Optional) A list of permission IDs to grant to the role.

## Attributes Reference

The following attributes are exported:

-   `user_count`: The number of users that have this role.

## Import

Roles can be imported using their ID, e.g.

```
$ terraform import datadog_role.example_role 000000-0000-0000-0000-000000000000
```
