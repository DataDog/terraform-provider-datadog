
# Create new integration_gcp_sts_account resource


resource "datadog_integration_gcp_sts_account" "foo" {
  automute        = "UPDATE ME"
  client_email    = "datadog-service-account@test-project.iam.gserviceaccount.com"
  host_filters    = "UPDATE ME"
  is_cspm_enabled = "UPDATE ME"
}