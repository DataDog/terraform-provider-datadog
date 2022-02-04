resource "datadog_logs_index_order" "sample_index_order" {
  name = "sample_index_order"

  indexes = [
    datadog_logs_index.sample_index.id
  ]
}
