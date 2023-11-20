# Create APM retention filter
resource "datadog_apm_retention_filter" "foo" {
  name = "Sample order"
  rate = "1.0"
  filter {
    query = "*"
  }
  filter_type = "spans-sampling-processor"
  enabled     = false
}

# Create APM reention filter order
resource "datadog_apm_retention_filter_order" "bar" {
  filter_ids = [datadog_apm_retention_filter.foo.id]
}