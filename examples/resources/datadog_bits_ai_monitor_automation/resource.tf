resource "datadog_bits_ai_monitor_automation" "example" {
  monitor_id = datadog_monitor.example.id
  enabled    = true
}
