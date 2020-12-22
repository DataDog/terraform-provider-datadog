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

- **email** (String, Required)

### Optional

- **access_role** (String, Optional)
- **disabled** (Boolean, Optional)
- **handle** (String, Optional, Deprecated)
- **id** (String, Optional) The ID of this resource.
- **is_admin** (Boolean, Optional, Deprecated)
- **name** (String, Optional)
- **role** (String, Optional, Deprecated)
- **roles** (Set of String, Optional)

### Read-only

- **verified** (Boolean, Read-only)

## Import

Import is supported using the following syntax:

```shell
# users can be imported using their ID, e.g.
terraform import datadog_user.example_user 6f1b44c0-30b2-11eb-86bc-279f7c1ebaa4
```
