resource "datadog_monitor_config_policy" "test" {
  policy_type = "tag"
  tag_policy {
    tag_key          = "env"
    tag_key_required = false
    valid_tag_values = ["staging", "prod"]
  }
}