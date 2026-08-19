# Create new rum_exclusion_filter resource

resource "datadog_rum_exclusion_filter" "testing_rum_exclusion_filter" {
  application_id = "<APPLICATION_ID>"
  name           = "testing.rum.exclusion_filter"
  event_type     = "error"
  query          = "@error.message:*extension*"
  enabled        = true
}
