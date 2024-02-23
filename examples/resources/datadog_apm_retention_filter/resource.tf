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
