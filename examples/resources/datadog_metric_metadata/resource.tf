# Manage a Datadog metric's metadata
resource "datadog_metric_metadata" "request_time" {
  metric      = "request.time"
  short_name  = "Request time"
  description = "99th percentile request time in milliseconds"
  type        = "gauge"
  unit        = "millisecond"
}
