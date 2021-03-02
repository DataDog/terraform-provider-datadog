# Manage a Datadog metric's metadata
resource "datadog_metric_tag_configuration" "testing_metric_tag_config" {
  # this distribution metric is the same as the import cassette
  metric_name         = "tf_TestAccDatadogMetricTagConfiguration_import_local_1614719241"
  metric_type         = "distribution"
  tags                = ["sport", "datacenter"]
  include_percentiles = true
}
