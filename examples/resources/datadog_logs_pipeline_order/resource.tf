resource "datadog_logs_pipeline_order" "sample_pipeline_order" {
  name = "sample_pipeline_order"

  pipelines = [
    datadog_logs_custom_pipeline.sample_pipeline.id,
    datadog_logs_integration_pipeline.python.id
  ]
}

