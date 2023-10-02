# Create new integration_gcp_sts resource

// Service account should have compute.viewer, monitoring.viewer, and cloudasset.viewer roles.
resource "google_service_account" "datadog_integration" {
  account_id   = "datadogintegration"
  display_name = "Datadog Integration"
  project      = "gcp-project"
}

// Grant token creator role to the Datadog principal account.
resource "google_service_account_iam_member" "sa_iam" {
  service_account_id = google_service_account.datadog_integration.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = format("serviceAccount:%s", datadog_integration_gcp_sts.foo.delegate_account_email)
}

resource "datadog_integration_gcp_sts" "foo" {
  client_email    = google_service_account.datadog_integration.email
  host_filters    = ["filter_one", "filter_two"]
  automute        = true
  is_cspm_enabled = false
}
