# Create new rum_metrics resource

resource "datadog_rum_metrics" "foo" {
  compute {
    aggregation_type    = "distribution"
    include_percentiles = True
    path                = "@duration"
  }
  event_type = "session"
  filter {
    query = "@service:web-ui: "
  }
  group_by {
    path     = "@browser.name"
    tag_name = "browser_name"
  }
  uniqueness {
    when = "match"
  }
}