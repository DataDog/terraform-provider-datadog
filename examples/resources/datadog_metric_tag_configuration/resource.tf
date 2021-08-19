# Manage a tag configuration for a Datadog distribution metric with/without percentiles
resource "datadog_metric_tag_configuration" "example_dist_metric" {
  metric_name         = "example.terraform.dist.metric"
  metric_type         = "distribution"
  tags                = ["sport", "datacenter"]
  include_percentiles = false
}

# Manage tag configurations for a Datadog count or gauge metric
resource "datadog_metric_tag_configuration" "example_count_metric" {
  metric_name  = "example.terraform.count.metric"
  metric_type  = "count"
  tags         = ["sport", "datacenter"]
  aggregations = [{ "time" : "avg", "space" : "min" }, { "time" : "avg", "space" : "max" }]
}
