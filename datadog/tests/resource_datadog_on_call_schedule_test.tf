resource "datadog_user" "foo" {
  email = "USER_EMAIL"
}

resource "datadog_team" "foo" {
  description = "Description"
  handle      = "TEAM_HANDLE"
  name        = "TEAM_NAME"
}

resource "datadog_on_call_schedule" "single_layer" {
  name      = "SCHEDULE_NAME"
  time_zone = "America/New_York"
  teams     = [datadog_team.foo.id]
  layer {
    effective_date = "EFFECTIVE_DATE"
    end_date       = "END_DATE"
    interval {
      days    = 1
      seconds = 300
    }
    rotation_start = "ROTATION_START"
    users          = [datadog_user.foo.id, null]
    name           = "Primary On-Call Layer"
    restriction {
      end_day    = "monday"
      end_time   = "17:00:00"
      start_day  = "monday"
      start_time = "09:00:00"
    }
  }
}
