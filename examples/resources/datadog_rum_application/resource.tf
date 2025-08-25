resource "datadog_rum_application" "rum_application" {
  name                              = "my-application"
  type                              = "browser"
  rum_event_processing_state        = "ALL"
  product_analytics_retention_state = "NONE"
}
