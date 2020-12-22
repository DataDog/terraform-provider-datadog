# Create a new Datadog - Google Cloud Platform integration
resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id     = "awesome-project-id"
  private_key_id = "1234567890123456789012345678901234567890"
  private_key    = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  client_email   = "awesome-service-account@awesome-project-id.iam.gserviceaccount.com"
  client_id      = "123456789012345678901"
  host_filters   = "foo:bar,buzz:lightyear"
}

