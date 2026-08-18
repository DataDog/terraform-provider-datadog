# Configure the built-in "Unused API Keys" governance control
resource "datadog_governance_control" "unused_api_keys" {
  detection_type = "unused_api_keys"

  detection_parameters = jsonencode({
    api_key_threshold = 30
  })

  notification_settings = [
    {
      event_type = "new_detection"
      enabled    = true
      targets = [
        {
          type   = "slack"
          handle = "#platform-governance"
        }
      ]
    }
  ]
}
