# Create new rum_retention_filter resource

resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
  app_id = "a7f5b0a3-9f73-4fdb-8fe1-91a8fc5a7a72"
  name = "testing.rum.retention_filter"
  event_type = "session"
  sample_rate = 41
  query = "@session.has_replay:true"
  enabled = false
}
