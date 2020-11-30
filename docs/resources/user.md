---
page_title: "datadog_user"
---

# datadog_user Resource

Provides a Datadog user resource. This can be used to create and manage Datadog users.

## Example Usage

```hcl
# Source a role
data "datadog_role" "ro_role" {
  filter = "Datadog Read Only Role"
}

# Create a new Datadog user
resource "datadog_user" "foo" {
  email  = "new@example.com"

  roles = [data.datadog_role.ro_role.id]
}
```

## Argument Reference

The following arguments are supported:

-   `email`: (Required) Email address for user
-   `name`: (Optional) Name for user
-   `roles`: (Optional) A list a role IDs to assign to the user.
-   `disabled`: (Optional) Whether the user is disabled
-   `handle`: (Deprecated) The user handle, must be a valid email.
-   `is_admin`: (Deprecated) (Optional) Whether the user is an administrator. **Warning**: the corresponding query parameter is ignored by the Datadog API, thus the argument would always trigger an execution plan.
-   `role`: (Deprecated) Role description for user. **Warning**: the corresponding query parameter is ignored by the Datadog API, thus the argument would always trigger an execution plan.
-   `access_role`: (Deprecated) Role description for user. Can be `st` (standard user), `adm` (admin user) or `ro` (read-only user). Default is `st`. `access_role` is ignored for new users created with this resource. New users have to use the `roles` attribute.

## Attributes Reference

The following attributes are exported:

-   `disabled`: Returns true if Datadog user is disabled (NOTE: Datadog does not actually delete users so this will be true for those as well)
-   `id`: ID of the Datadog user
-   `verified`: Returns true if Datadog user is verified

## Import

users can be imported using their ID, e.g.

```
$ terraform import datadog_user.example_user 6f1b44c0-30b2-11eb-86bc-279f7c1ebaa4
```
