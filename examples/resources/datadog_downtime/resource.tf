# Example: downtime for a specific monitor
# Create a new daily 1700-0900 Datadog downtime for a specific monitor id
resource "datadog_downtime" "foo" {
  scope = ["*"]
  start = 1483308000
  end   = 1483365600
  monitor_id = 12345

  recurrence {
    type   = "days"
    period = 1
  }
}

# Example: downtime for all monitors
# Create a new daily 1700-0900 Datadog downtime for all monitors
resource "datadog_downtime" "foo" {
  scope = ["*"]
  start = 1483308000
  end   = 1483365600

  recurrence {
    type   = "days"
    period = 1
  }
}
