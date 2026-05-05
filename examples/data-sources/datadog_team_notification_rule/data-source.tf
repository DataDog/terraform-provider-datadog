# Get a specific notification rule for a team
data "datadog_team_notification_rule" "example" {
  team_id = "00000000-0000-0000-0000-000000000000"
  rule_id = "11111111-1111-1111-1111-111111111111"
}
