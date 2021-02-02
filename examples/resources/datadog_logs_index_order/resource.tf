resource "datadog_logs_index_order" "sample_index_order" {
  name = "sample_index_order"
  depends_on = [
    "datadog_logs_index.sample_index"
  ]
  indexes = [
    "${datadog_logs_index.sample_index.id}"
  ]
}
