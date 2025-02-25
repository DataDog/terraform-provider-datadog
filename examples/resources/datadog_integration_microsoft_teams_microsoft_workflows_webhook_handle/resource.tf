# Create a new integration_ms_teams_workflows_webhook_handle resource

resource "datadog_integration_ms_teams_workflows_webhook_handle" "testing_microsoft_workflows_webhook_handle" {
  name = "sample_handle_name"
  url  = "https://fake.url.com"
}
