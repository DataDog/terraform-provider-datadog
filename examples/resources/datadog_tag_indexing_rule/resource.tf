resource "datadog_tag_indexing_rule" "example" {
  name                = "Index env and service tags for all web metrics"
  metric_name_matches = ["web.*", "http.*"]
  tags                = ["env", "service", "version"]
  exclude_tags_mode   = false

  options {
    version = 1
    data {
      manage_preexisting_metrics = true
      override_previous_rules    = false
    }
  }
}
