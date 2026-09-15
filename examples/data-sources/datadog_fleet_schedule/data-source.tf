# Reading Fleet Automation schedules requires the hosts_read permission.
data "datadog_fleet_schedule" "nonproduction" {
  id = "<schedule_id>"
}
