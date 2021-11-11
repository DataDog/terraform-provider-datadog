# Create a new Datadog webhook custom variable.

resource "datadog_webhook_custom_variable" "foo" {
  name      = "EXAMPLE_VARIABLE"
  value     = "EXAMPLE-VALUE"
  is_secret = true
}