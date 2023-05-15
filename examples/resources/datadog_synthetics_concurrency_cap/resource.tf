# Example Usage (Synthetics Concurrency Cap Configuration)
resource "datadog_synthetics_concurrency_cap" "this" {
  on_demand_concurrency_cap = 1
}