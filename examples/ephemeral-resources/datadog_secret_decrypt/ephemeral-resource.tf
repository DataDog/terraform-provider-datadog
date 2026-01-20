# Decrypt an encrypted API key for use with a secret manager
# Requires Terraform 1.11+

variable "encryption_key" {
  type        = string
  description = "32-byte encryption key"
  sensitive   = true
}

resource "datadog_api_key" "example" {
  name              = "my-api-key"
  encryption_key_wo = var.encryption_key
}

# Only decrypt when encrypted_key is available
# After secret is stored in Secrets Manager, you can remove encryption_key_wo
# and this block will be skipped on future runs
ephemeral "datadog_secret_decrypt" "api_key" {
  count             = datadog_api_key.example.encrypted_key != null ? 1 : 0
  ciphertext        = datadog_api_key.example.encrypted_key
  encryption_key_wo = var.encryption_key
}

resource "aws_secretsmanager_secret" "datadog_api_key" {
  name = "datadog-api-key"
}

# Use try() to handle the case where decrypter wasn't created
resource "aws_secretsmanager_secret_version" "api_key" {
  secret_id        = aws_secretsmanager_secret.datadog_api_key.id
  secret_string_wo = try(ephemeral.datadog_secret_decrypt.api_key[0].value, "")
}
