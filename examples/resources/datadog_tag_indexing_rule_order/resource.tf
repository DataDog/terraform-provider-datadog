resource "datadog_tag_indexing_rule" "broad" {
  name                = "Broad rule applied first"
  metric_name_matches = ["*"]
  tags                = ["env", "service"]
  exclude_tags_mode   = false
}

resource "datadog_tag_indexing_rule" "specific" {
  name                = "Specific override for web metrics"
  metric_name_matches = ["web.*"]
  tags                = ["env", "service", "version", "host"]
  exclude_tags_mode   = false
}

# Enforce evaluation order: broad rule first, then specific override.
# rule_ids must list EVERY active tag indexing rule in the org (this resource owns the whole-org
# order). Any rule omitted here will be rejected by the API.
resource "datadog_tag_indexing_rule_order" "example" {
  name = "main"
  rule_ids = [
    datadog_tag_indexing_rule.broad.id,
    datadog_tag_indexing_rule.specific.id,
  ]
}
