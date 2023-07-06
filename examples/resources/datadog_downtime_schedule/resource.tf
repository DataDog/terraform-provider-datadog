
# Create new downtime_schedule resource


resource "datadog_downtime_schedule" "foo" {
  display_timezone = "America/New_York"
  message          = "Message about the downtime"
  monitor_identifier {
  }
  mute_first_recovery_notification = "UPDATE ME"
  notify_end_states                = ["alert", "warn"]
  notify_end_types                 = ["canceled", "expired"]
  schedule {
  }
  scope = "env:(staging OR prod) AND datacenter:us-east-1"
}