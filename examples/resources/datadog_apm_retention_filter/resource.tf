# Create APM retention filter
resource "datadog_apm_retention_filter" "foo" {
  name = "Sample order"
  rate = "1.0"
  filter {
    query = "service:sample AND env:production AND @http.method:GET AND app:sampleapp AND @http.status_code:200 AND @duration:>600000000"
  }
  filter_type = "spans-sampling-processor"
  enabled     = false
}
