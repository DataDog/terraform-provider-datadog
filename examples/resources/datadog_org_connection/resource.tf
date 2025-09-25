
# Create new org connection resource

resource "datadog_org_connection" "foo" {
  connection_types = [
    "metrics",
    "logs"
  ]
  sink_org_id = "00000000-0000-0000-0000-000000000000"
}
