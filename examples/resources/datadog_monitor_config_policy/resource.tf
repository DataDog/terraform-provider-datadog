resource "datadog_monitor_config_policy" "test" {
  policy_type = "tag"
  tag_policy {
    tag_key          = "env"
    tag_key_required = false
    valid_tag_values = ["staging", "prod"]
  }
}

resource "datadog_monitor_config_policy" "downtime_example" {
  policy_type = "downtime"
  downtime_policy {
    max_duration_ms = 3600000
  }
}
