resource "datadog_rum_application" "rum_application" {
  name                              = "my-application"
  type                              = "browser"
  rum_event_processing_state        = "ALL"
  product_analytics_retention_state = "NONE"
}

# Create the application without the default retention filters to manage them yourself.
resource "datadog_rum_application" "rum_application_custom_filters" {
  name                             = "my-application-custom-filters"
  type                             = "browser"
  create_default_retention_filters = false
}
