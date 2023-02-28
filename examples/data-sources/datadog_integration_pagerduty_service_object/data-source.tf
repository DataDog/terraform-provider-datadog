data "datadog_integration_pagerduty_service_object" "foo" {
  service_name = "foo"
}

resource "datadog_integration_pagerduty_service_object" "foo" {
  count        = length(datadog_integration_pagerduty_service_object.foo) > 0 ? 1 : 0
  service_name = "foo"
  service_key  = "foo"
}