# Decrypt an encrypted API key for use with a secret manager
# Requires Terraform 1.11+

variable "encryption_key" {
  type = string
}

resource "datadog_api_key" "example" {
  name              = "my-api-key"
  encryption_key_wo = var.encryption_key
}

ephemeral "datadog_secret_decrypt" "api_key" {
  ciphertext        = datadog_api_key.example.encrypted_key
  encryption_key_wo = var.encryption_key
}

# Use the decrypted value with a secret manager
resource "aws_secretsmanager_secret_version" "api_key" {
  secret_id        = aws_secretsmanager_secret.datadog_api_key.id
  secret_string_wo = ephemeral.datadog_secret_decrypt.api_key.value
}
