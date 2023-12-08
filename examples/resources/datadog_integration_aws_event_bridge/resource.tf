# Create new integration_aws_event_bridge resource

resource "datadog_integration_aws_event_bridge" "foo" {
  body {
    account_id           = "123456789012"
    create_event_bus     = True
    event_generator_name = "app-alerts"
    region               = "us-east-1"
  }
}