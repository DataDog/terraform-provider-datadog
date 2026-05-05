# Get all notification rules for a team
data "datadog_team_notification_rules" "example" {
  team_id = "00000000-0000-0000-0000-000000000000"
}
