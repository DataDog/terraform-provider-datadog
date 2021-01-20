---
page_title: "datadog_user Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog user resource. This can be used to create and manage Datadog users.
---

# Resource `datadog_user`

Provides a Datadog user resource. This can be used to create and manage Datadog users.

## Example Usage

```terraform
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

## Schema

### Required

- **email** (String) Email address for user.

### Optional

- **access_role** (String, Deprecated) Role description for user. Can be `st` (standard user), `adm` (admin user) or `ro` (read-only user). Default is `st`. `access_role` is ignored for new users created with this resource. New users have to use the `roles` attribute.
- **disabled** (Boolean) Whether the user is disabled.
- **handle** (String, Deprecated) The user handle, must be a valid email.
- **id** (String) The ID of this resource.
- **is_admin** (Boolean, Deprecated) Whether the user is an administrator. Warning: the corresponding query parameter is ignored by the Datadog API, thus the argument would always trigger an execution plan.
- **name** (String) Name for user.
- **role** (String, Deprecated) Role description for user. Warning: the corresponding query parameter is ignored by the Datadog API, thus the argument would always trigger an execution plan.
- **roles** (Set of String) A list a role IDs to assign to the user.
- **send_user_invitation** (Boolean) Whether an invitation email should be sent when the user is created.

### Read-only

- **user_invitation_id** (String) The ID of the user invitation that was sent when creating the user.
- **verified** (Boolean) Returns true if Datadog user is verified.

## Import

Import is supported using the following syntax:

```shell
terraform import datadog_user.example_user 6f1b44c0-30b2-11eb-86bc-279f7c1ebaa4
```
