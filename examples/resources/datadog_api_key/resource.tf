# Create a new Datadog API Key
resource "datadog_api_key" "foo" {
  name = "foo-application"
}

# Create an API Key with encryption for secure secret management
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
