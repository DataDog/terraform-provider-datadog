
# Create new downtime_schedule resource


resource "datadog_downtime_schedule" "downtime_schedule_example" {
  scope = "env:us9-prod7 AND team:test123"
  monitor_identifier {
    monitor_tags = ["test:123", "data:test"]
  }
  recurring_schedule {
    recurrence {
      duration = "1h"
      rrule    = "FREQ=DAILY;INTERVAL=1"
      start    = "2050-01-02T03:04:05"
    }
    timezone = "America/New_York"
  }
  display_timezone                 = "America/New_York"
  message                          = "Message about the downtime"
  mute_first_recovery_notification = true
  notify_end_states                = ["alert", "warn"]
  notify_end_types                 = ["canceled", "expired"]
}
