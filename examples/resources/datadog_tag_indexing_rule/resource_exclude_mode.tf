resource "datadog_tag_indexing_rule" "exclude_example" {
  name                = "Exclude unused tags from all web metrics"
  metric_name_matches = ["web.*", "http.*"]
  tags                = ["debug_id", "internal_trace_id"]
  exclude_tags_mode   = true

  options = {
    version = 1
    data = {
      dynamic_tags = {
        exclude_not_queried_window_seconds = 604800
        exclude_not_used_in_assets         = true
      }
    }
  }
}
