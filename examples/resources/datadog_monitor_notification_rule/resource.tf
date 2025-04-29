# Create new monitor_notification_rule resource

resource "datadog_monitor_notification_rule" "foo" {
  name       = "A notification rule name"
  recipients = ["slack-test-channel", "jira-test"]
  filter {
    tags = ["env:foo"]
  }
}