# Create a new Datadog API Key
resource "datadog_api_key" "foo" {
  name = "foo-application"
}

# Create an API Key with encryption for secure secret management
# Requires Terraform 1.11+
ephemeral "random_password" "encryption_key" {
  length = 32
}

resource "datadog_api_key" "encrypted" {
  name              = "encrypted-api-key"
  encryption_key_wo = ephemeral.random_password.encryption_key.result
}

# Decrypt the key to pass to a secret manager
ephemeral "datadog_secret_decrypt" "api_key" {
  ciphertext        = datadog_api_key.encrypted.encrypted_key
  encryption_key_wo = ephemeral.random_password.encryption_key.result
}

# Store the decrypted key in AWS Secrets Manager (using write-only attribute)
resource "aws_secretsmanager_secret_version" "api_key" {
  secret_id        = aws_secretsmanager_secret.datadog_api_key.id
  secret_string_wo = ephemeral.datadog_secret_decrypt.api_key.value
}
