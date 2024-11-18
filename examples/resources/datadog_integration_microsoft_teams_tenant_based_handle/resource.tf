# Create a new integration_ms_teams_tenant_based_handle resource

resource "datadog_integration_ms_teams_tenant_based_handle" "testing_tenant_based_handle" {
  name         = "sample_handle_name"
  tenant_name  = "sample_tenant_name"
  team_name    = "sample_team_name"
  channel_name = "sample_channel_name"
}
