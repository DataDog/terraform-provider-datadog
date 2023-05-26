# Create new integration_gcp_sts_account resource

resource "datadog_integration_gcp_sts_account" "foo" {
  client_email    = "service-account@example.com"
  host_filters    = ["filter_one", "filter_two"]
  automute        = true
  is_cspm_enabled = true
}