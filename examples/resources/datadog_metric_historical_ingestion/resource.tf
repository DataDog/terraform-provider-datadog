# The metric must already exist and be eligible for Historical Metrics Ingestion.
resource "datadog_metric_historical_ingestion" "example" {
  metric_name = "example.application.requests"
}
