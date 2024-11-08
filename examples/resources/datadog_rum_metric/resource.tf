# Create new rum_metric resource

resource "datadog_rum_metric" "testing_rum_metric" {
  name = "testing.rum.metric"
  compute {
    aggregation_type    = "distribution"
    include_percentiles = true
    path                = "@duration"
  }
  event_type = "session"
  filter {
    query = "@service:web-ui"
  }
  group_by {
    path     = "@browser.name"
    tag_name = "browser_name"
  }
  uniqueness {
    when = "match"
  }
}