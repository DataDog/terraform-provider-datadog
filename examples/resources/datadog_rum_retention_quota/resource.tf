# Create new rum_retention_quota resource

resource "datadog_rum_retention_quota" "testing_rum_retention_quota" {
  application_id = "<APPLICATION_ID>"
  mode           = "custom"
  custom {
    window_type          = "daily"
    session_limit        = 10000
    daily_reset_time     = "00:00"
    daily_reset_timezone = "+00:00"
    quota_reached_action = "stop"
  }
}
