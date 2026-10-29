# Create a derived metric that combines two metric queries
resource "datadog_derived_metric" "cpu_total" {
  name    = "custom.cpu.total"
  formula = "a + b"
  queries = [
    {
      name        = "a"
      data_source = "metrics"
      query       = "avg:system.cpu.user{*}"
    },
    {
      name        = "b"
      data_source = "metrics"
      query       = "avg:system.cpu.system{*}"
    },
  ]
}
