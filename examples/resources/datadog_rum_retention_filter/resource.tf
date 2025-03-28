# Create new rum_retention_filter resource

resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
  application_id = "<APPLICATION_ID>"
  name           = "testing.rum.retention_filter"
  event_type     = "session"
  sample_rate    = 41
  query          = "@session.has_replay:true"
  enabled        = false
}
