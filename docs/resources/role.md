---
page_title: "datadog_role Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog role resource. This can be used to create and manage Datadog roles.
---

# Resource `datadog_role`

Provides a Datadog role resource. This can be used to create and manage Datadog roles.



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


