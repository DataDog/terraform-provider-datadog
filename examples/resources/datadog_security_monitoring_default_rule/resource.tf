resource "datadog_security_monitoring_default_rule" "adefaultrule" {
  enabled = true

  # Change the notifications for the high case
  case {
    status        = "high"
    notifications = ["@me"]
  }
}
