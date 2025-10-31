# Example: Using ephemeral resources for enhanced security
# Set store_sensitive_state = false in your provider configuration

terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
    }
  }
}

provider "datadog" {
  # Enhanced security: API key values won't be stored in state
  store_sensitive_state = false
}

# Create the API key resource (key value won't be stored in state)
resource "datadog_api_key" "example" {
  name = "Example API Key"
}

# Access the key value using ephemeral resource (not stored in state)
ephemeral "datadog_api_key" "example" {
  id = datadog_api_key.example.id
}

# Use the ephemeral key value in other resources
resource "some_external_resource" "example" {
  api_key = ephemeral.datadog_api_key.example.key
}

# Or store in locals for reuse
locals {
  api_key = ephemeral.datadog_api_key.example.key
}