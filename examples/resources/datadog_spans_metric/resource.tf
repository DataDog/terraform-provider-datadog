# Create new spans_metric resource

resource "datadog_spans_metric" "testing_spans_metric" {
  name = "testing.span.metric"
  compute {
    aggregation_type    = "distribution"
    include_percentiles = false
    path                = "@duration"
  }
  filter {
    query = "@http.status_code:200 service:my-service"
  }
  group_by {
    path     = "resource_name"
    tag_name = "resource_name"
  }
}
