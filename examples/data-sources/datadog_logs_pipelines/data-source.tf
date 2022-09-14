# Gets all pipelines
data "datadog_logs_pipelines" "pipelines" {}

# Using data source to set pipeline order
resource "datadog_logs_pipeline_order" "lpo" {
  name = "lpo"
  pipelines = [for pipeline in
  (data.datadog_logs_pipelines.pipelines.logs_pipelines) : pipeline.id]
}
