# Fleet schedule mutations use Preview API endpoints. The application key must
# have both the agent_upgrade_write and hosts_read permissions.
resource "datadog_fleet_schedule" "nonproduction" {
  name              = "Non-production Agent upgrades"
  query             = "env:staging service:example"
  status            = "inactive"
  version_to_latest = 1

  rule = {
    days_of_week                = ["Tue", "Thu"]
    maintenance_window_duration = 120
    start_maintenance_window    = "03:00"
    timezone                    = "America/New_York"
  }
}
