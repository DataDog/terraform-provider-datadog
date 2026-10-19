resource "datadog_tag_indexing_rule_exemption" "example" {
  metric_name = "system.cpu.user"
  reason      = "High-cardinality metric; exempted to reduce index costs"
}
