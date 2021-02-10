resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  service_name = "testing_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  service_name = "testing_bar"
  service_key  = "54321098765432109876"
}
