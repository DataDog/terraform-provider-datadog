# The metric must already exist. This resource manages tags only;
# it does not enable or disable distribution percentiles.
resource "datadog_metric_tag_configuration_v2" "example" {
  metric_name = "app.request.duration"
  metric_type = "distribution"
  tags        = ["env", "service"]
}
