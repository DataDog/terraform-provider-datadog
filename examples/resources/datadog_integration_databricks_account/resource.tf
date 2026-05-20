# Create a Databricks integration account using OAuth (recommended).
resource "datadog_integration_databricks_account" "oauth_example" {
  name          = "databricks-prod"
  workspace_url = "https://your-workspace.cloud.databricks.com"

  auth_config {
    oauth {
      client_id             = "abc123def456"
      client_secret         = "secret-value"
      databricks_account_id = "11111111-2222-3333-4444-555555555555"
    }
  }

  djm_enabled             = true
  serverless_jobs_enabled = true
}

# Alternative: Databricks integration account using a Personal Access Token (PAT).
# OAuth is preferred for new deployments; PAT is kept for backwards compatibility.
resource "datadog_integration_databricks_account" "pat_example" {
  name          = "databricks-legacy"
  workspace_url = "https://your-workspace.cloud.databricks.com"

  auth_config {
    pat {
      token = "dapi-..."
    }
  }
}
