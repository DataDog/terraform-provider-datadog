# Manage a Datadog metric's metadata
resource "datadog_metric_tag_configuration" "testing_metric_tag_config" {
  # TODO[efraese] use an id for a distribution metric
  id                  = ""
  metric_type         = "distribution"
  tags                = ["sport", "datacenter"]
  include_percentiles = true
}
