---
page_title: "datadog_user"
---

# datadog_user Resource

Provides a Datadog user resource. This can be used to create and manage Datadog users.

## Example Usage

```hcl
# Create a new Datadog user
resource "datadog_user" "foo" {
  email  = "new@example.com"
  handle = "new@example.com"
  name   = "New User"
}
```

## Argument Reference

The following arguments are supported:

-   `access_role`: (Optional) Role description for user. Can be `st` (standard user), `adm` (admin user) or `ro` (read-only user). Default is `st`.
-   `disabled`: (Optional) Whether the user is disabled
-   `email`: (Required) Email address for user
-   `handle`: (Required) The user handle, must be a valid email.
-   `is_admin`: (Deprecated) (Optional) Whether the user is an administrator. **Warning**: the corresponding query parameter is ignored by the Datadog API, thus the argument would always trigger an execution plan.
-   `name`: (Required) Name for user
-   `role`: (Deprecated) Role description for user. **Warning**: the corresponding query parameter is ignored by the Datadog API, thus the argument would always trigger an execution plan.

## Attributes Reference

The following attributes are exported:

-   `disabled`: Returns true if Datadog user is disabled (NOTE: Datadog does not actually delete users so this will be true for those as well)
-   `id`: ID of the Datadog user
-   `verified`: Returns true if Datadog user is verified

## Import

users can be imported using their handle, e.g.

```
$ terraform import datadog_user.example_user existing@example.com
```
