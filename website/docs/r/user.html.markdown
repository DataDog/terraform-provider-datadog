---
layout: "datadog"
page_title: "Datadog: datadog_user"
sidebar_current: "docs-datadog-resource-user"
description: |-
  Provides a Datadog user resource. This can be used to create and manage users.
---

# datadog_user

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

* `disabled` - (Optional) Whether the user is disabled
* `email` - (Required) Email address for user
* `handle` - (Required) The user handle, must be a valid email.
* `is_admin` - (Deprecated) (Optional) Whether the user is an administrator
* `name` - (Required) Name for user
* `role` - (Deprecated) Role description for user. **Warning**: the corresponding query parameter is ignored by the Datadog API, thus the argument would always trigger an execution plan.

## Attributes Reference

The following attributes are exported:

* `disabled` - Returns true if Datadog user is disabled (NOTE: Datadog does not actually delete users so this will be true for those as well)
* `id` - ID of the Datadog user
* `verified` - Returns true if Datadog user is verified

## Import

users can be imported using their handle, e.g.

```
$ terraform import datadog_user.example_user existing@example.com
```
