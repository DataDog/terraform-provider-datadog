# Basic Usage
resource "datadog_synthetics_global_variable" "test_variable" {
  name        = "EXAMPLE_VARIABLE"
  description = "Description of the variable"
  tags        = ["foo:bar", "env:test"]
  value       = "variable-value"
}

# Write-Only Value (Recommended for Terraform 1.11+)
resource "datadog_synthetics_global_variable" "secure_variable" {
  name        = "SECURE_VARIABLE"
  description = "Secure global variable with write-only value"
  tags        = ["foo:bar", "env:production"]
  secure      = true

  # Write-only value with version trigger
  value_wo         = var.secret_value
  value_wo_version = "1" # Any string: "1", "v2.1", "2024-Q1", etc.
}

# Advanced: Automated Version Management
locals {
  secret_keepers = {
    rotation_date   = "2024-02-15"
    environment     = "production"
    security_policy = "v3.1"
  }

  # Auto-generate version from keepers
  secret_version = "rotation-${substr(md5(jsonencode(local.secret_keepers)), 0, 8)}"
}

resource "datadog_synthetics_global_variable" "automated_rotation" {
  name        = "AUTO_ROTATED_VARIABLE"
  description = "Variable with automated rotation"
  tags        = ["foo:bar", "env:production"]
  secure      = true

  # Version automatically updates when any keeper changes
  value_wo         = var.secret_value
  value_wo_version = local.secret_version
}
