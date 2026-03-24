# Manage a Datadog log restriction query
resource "datadog_logs_restriction_query" "test_lrq" {
  restriction_query = "service:foo"

  role_ids = [
    "00000000-0000-1111-0000-000000000000",
    "11111111-1111-0000-1111-111111111111",
  ]
}
