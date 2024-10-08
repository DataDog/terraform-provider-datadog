---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_service_account Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve information about an existing Datadog service account.
---

# datadog_service_account (Data Source)

Use this data source to retrieve information about an existing Datadog service account.



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `exact_match` (Boolean) When true, `filter` string is exact matched against the user's `email`, followed by `name` attribute.
- `filter` (String) Filter all users and service accounts by name, email, or role.
- `filter_status` (String) Filter on status attribute. Comma separated list, with possible values `Active`, `Pending`, and `Disabled`.
- `id` (String) The service account's ID.

### Read-Only

- `disabled` (Boolean) Whether the user is disabled.
- `email` (String) Email of the user.
- `handle` (String) Handle of the user.
- `icon` (String) URL of the user's icon.
- `name` (String) Name of the user.
- `roles` (Set of String) Roles assigned to this service account.
- `status` (String) Status of the user.
- `title` (String) Title of the user.
- `verified` (Boolean) Whether the user is verified.
