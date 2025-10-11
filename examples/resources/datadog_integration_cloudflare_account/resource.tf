# Basic Usage
resource "datadog_integration_cloudflare_account" "basic" {
  api_key = "12345678910abc"
  email   = "test-email@example.com"
  name    = "test-name"
}

# Write-Only API Key (Recommended for Terraform 1.11+)
resource "datadog_integration_cloudflare_account" "secure" {
  name  = "prod-cloudflare"
  email = "admin@company.com"

  # Write-only API key with version trigger
  api_key_wo         = var.cloudflare_api_key
  api_key_wo_version = "1" # Any string: "1", "v2.1", "2024-Q1", etc.
}

# Advanced: Automated Version Management
locals {
  cloudflare_keepers = {
    rotation_date   = "2024-02-15"
    environment     = "production"
    security_policy = "v3.1"
  }

  # Auto-generate version from keepers
  api_key_version = "rotation-${substr(md5(jsonencode(local.cloudflare_keepers)), 0, 8)}"
}

resource "datadog_integration_cloudflare_account" "automated" {
  name  = "prod-cloudflare"
  email = "admin@company.com"

  # Version automatically updates when any keeper changes
  api_key_wo         = var.cloudflare_api_key
  api_key_wo_version = local.api_key_version
}
