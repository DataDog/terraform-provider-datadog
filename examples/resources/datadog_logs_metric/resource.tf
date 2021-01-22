resource "datadog_logs_metric" "testing_logs_metric" {
  name = "testing.logs.metric"
  compute {
    aggregation_type = "distribution"
    path             = "@duration"
  }
  filter {
    query = "service:test"
  }
  group_by {
    path     = "@status"
    tag_name = "status"
  }
}
