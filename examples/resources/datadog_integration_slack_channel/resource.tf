resource "datadog_integration_slack_channel" "test_channel" {
  account_name = "foo"
  channel_name = "#test_channel"

  display {
    message  = true
    notified = false
    snapshot = false
    tags     = true
  }
}
