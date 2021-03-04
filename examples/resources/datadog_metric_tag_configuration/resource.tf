# Manage a Datadog metric tag configuration
resource "datadog_metric_tag_configuration" "testing_metric_tag_config" {
  metric_name         = "example.terraform.metric.name"
  metric_type         = "distribution"
  tags                = ["sport", "datacenter"]
  include_percentiles = false
}
