# List the Slack channels configured for a Slack integration account
data "datadog_integration_slack_channels" "example" {
  account_name = "my_slack_account"
}
