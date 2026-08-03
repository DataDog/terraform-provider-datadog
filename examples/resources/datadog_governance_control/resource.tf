# Configure the built-in "Unused API Keys" governance control
resource "datadog_governance_control" "unused_api_keys" {
  detection_type      = "unused_api_keys"
  detection_frequency = "daily"

  notification_type      = "slack"
  notification_frequency = "daily"
  notification_parameters = jsonencode({
    slack_channel = "#platform-governance"
  })
}
