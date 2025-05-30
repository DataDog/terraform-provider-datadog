---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_users Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve information about existing users for use in other resources.
---

# datadog_users (Data Source)

Use this data source to retrieve information about existing users for use in other resources.

## Example Usage

```terraform
data "datadog_users" "test" {
  filter        = "user.name@company.com"
  filter_status = "Active,Pending"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `filter` (String) Filter all users by the given string.
- `filter_status` (String) Filter on status attribute. Comma-separated list with possible values of Active, Pending, and Disabled.

### Read-Only

- `id` (String) The ID of this resource.
- `users` (List of Object) List of users (see [below for nested schema](#nestedatt--users))

<a id="nestedatt--users"></a>
### Nested Schema for `users`

Read-Only:

- `created_at` (String)
- `disabled` (Boolean)
- `email` (String)
- `handle` (String)
- `icon` (String)
- `id` (String)
- `mfa_enabled` (Boolean)
- `modified_at` (String)
- `name` (String)
- `service_account` (Boolean)
- `status` (String)
- `title` (String)
- `verified` (Boolean)
