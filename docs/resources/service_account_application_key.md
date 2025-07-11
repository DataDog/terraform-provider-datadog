---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_service_account_application_key Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog service_account_application_key resource. This can be used to create and manage Datadog service account application keys.
---

# datadog_service_account_application_key (Resource)

Provides a Datadog `service_account_application_key` resource. This can be used to create and manage Datadog service account application keys.

## Example Usage

```terraform
# Create new service_account_application_key resource
resource "datadog_service_account_application_key" "foo" {
  service_account_id = "00000000-0000-1234-0000-000000000000"
  name               = "Application key for managing dashboards"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the application key.
- `service_account_id` (String) ID of the service account that owns this key.

### Optional

- `scopes` (Set of String) Authorization scopes for the Application Key. Application Keys configured with no scopes have full access.

### Read-Only

- `created_at` (String) Creation date of the application key.
- `id` (String) The ID of this resource.
- `key` (String, Sensitive) The value of the service account application key. This value cannot be imported.
- `last4` (String) The last four characters of the application key.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# Importing a service account's application key cannot import the value of the key.
terraform import datadog_service_account_application_key.this "<service_account_id>:<application_key_id>"
```
