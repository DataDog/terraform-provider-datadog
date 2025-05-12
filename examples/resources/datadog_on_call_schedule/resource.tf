# Create new on_call_schedule resource

resource "datadog_on_call_schedule" "test" {
  name      = "Team A On-Call"
  tags      = ["foo:bar"]
  time_zone = "America/New_York"
  teams = ["00000000-aba2-0000-0000-000000000000"]
  layers {
    name = "Primary On-Call Layer"
    effective_date = "2025-01-01T00:00:00Z"
    end_date       = "2026-01-01T00:00:00Z"
    rotation_start = "2025-01-01T00:00:00Z"
    interval {
      days    = 1
      seconds = 300
    }
    users = ["00000000-aba1-0000-0000-000000000000"]
    restrictions {
      end_day    = "monday"
      end_time   = "17:00:00"
      start_day  = "monday"
      start_time = "09:00:00"
    }
  }
}
