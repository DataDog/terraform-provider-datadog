resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  service_name = "testing_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  service_name = "testing_bar"
  service_key  = "54321098765432109876"
}

# With Terraform 1.11 or later, use the write-only attribute to keep the service key out of the state.
# Change service_key_wo_version to send a new service_key_wo value.
resource "datadog_integration_pagerduty_service_object" "testing_baz" {
  service_name           = "testing_baz"
  service_key_wo         = var.pagerduty_service_key
  service_key_wo_version = "1"
}
