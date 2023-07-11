resource "datadog_downtime_schedule" "downtime_schedule_test" {
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
  }
  message = "host migration"
}
