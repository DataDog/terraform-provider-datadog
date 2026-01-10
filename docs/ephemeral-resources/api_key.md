---
page_title: "datadog_api_key Ephemeral Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this ephemeral resource to retrieve a Datadog API key without storing it in Terraform state. This is the recommended approach for securely accessing API keys, for example to pass to a secrets manager.
---

# datadog_api_key (Ephemeral Resource)

Use this ephemeral resource to retrieve a Datadog API key without storing it in Terraform state. This is the recommended approach for securely accessing API keys, for example to pass to a secrets manager.

~> **Note:** Ephemeral resources require Terraform 1.10 or later. The API key value is retrieved at runtime and is never persisted to state or plan files.

## Example Usage

### Lookup by ID

```terraform
ephemeral "datadog_api_key" "example" {
  id = "abc123def456"
}
```

### Lookup by Name

```terraform
ephemeral "datadog_api_key" "example" {
  name        = "my-api-key"
  exact_match = true
}
```

### Pass to AWS Secrets Manager (using write-only argument)

```terraform
ephemeral "datadog_api_key" "example" {
  name        = "my-api-key"
  exact_match = true
}

resource "aws_secretsmanager_secret" "datadog" {
  name = "datadog-api-key"
}

# Requires AWS provider support for write-only arguments
resource "aws_secretsmanager_secret_version" "datadog" {
  secret_id        = aws_secretsmanager_secret.datadog.id
  secret_string_wo = ephemeral.datadog_api_key.example.key
}
```

### Pass to AWS Secrets Manager (using provisioner workaround)

If your AWS provider version doesn't support write-only arguments, you can use a provisioner:

```terraform
ephemeral "datadog_api_key" "example" {
  name        = "my-api-key"
  exact_match = true
}

resource "aws_secretsmanager_secret" "datadog" {
  name = "datadog-api-key"
}

resource "null_resource" "store_secret" {
  provisioner "local-exec" {
    command = "aws secretsmanager put-secret-value --secret-id ${aws_secretsmanager_secret.datadog.id} --secret-string '${ephemeral.datadog_api_key.example.key}'"
  }
}
```

## Schema

### Optional

- `exact_match` (Boolean) Whether to use exact match when searching by name.
- `id` (String) The ID of the API key.
- `name` (String) Name for API Key.

### Read-Only

- `key` (String, Sensitive) The value of the API Key. This value is never stored in Terraform state.
- `remote_config_read_enabled` (Boolean) Whether the API key is used for remote config.
