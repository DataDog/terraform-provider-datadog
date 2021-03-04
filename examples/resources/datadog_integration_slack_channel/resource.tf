resource "datadog_integration_slack_channel" "slack_channel" {
  display {
    message  = true
    notified = false
    snapshot = false
    tags     = true
  }
  channel_name = "#test_channel"
  account_name = "foo"
}
