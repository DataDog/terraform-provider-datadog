---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_integration_confluent_account Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog IntegrationConfluentAccount resource. This can be used to create and manage Datadog integration_confluent_account.
---

# datadog_integration_confluent_account (Resource)

Provides a Datadog IntegrationConfluentAccount resource. This can be used to create and manage Datadog integration_confluent_account.

## Example Usage

```terraform
# Create new integration_confluent_account resource

resource "datadog_integration_confluent_account" "foo" {
  api_key    = "TESTAPIKEY123"
  api_secret = "test-api-secret-123"
  tags       = ["mytag", "mytag2:myvalue"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `api_key` (String) The API key associated with your Confluent account.
- `api_secret` (String, Sensitive) The API secret associated with your Confluent account.

### Optional

- `tags` (Set of String) A list of strings representing tags. Can be a single key, or key-value pairs separated by a colon.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# Confluent account ID can be retrieved using the ListConfluentAccounts endpoint
# https://docs.datadoghq.com/api/latest/confluent-cloud/#list-confluent-accounts

terraform import datadog_integration_confluent_account.new_list "<ID>"
```
