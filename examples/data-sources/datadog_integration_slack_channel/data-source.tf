data "datadog_integration_slack_channel" "foo" {
  account_name = "test_account"
  channel_name = "foo"

}

resource "datadog_integration_slack_channel" "foo" {
  count        = length(datadog_integration_slack_channel.foo) > 0 ? 1 : 0
  account_name = "test_account"
  channel_name = "foo"
  display {
    message = false
    notified = false
    snapshot = false
    tags = false
  }
}