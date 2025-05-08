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
  tags      = ["foo:bar"]
  time_zone = "America/New_York"
  team_ids  = [datadog_team.foo.id]
  layer {
    effective_date = "EFFECTIVE_DATE"
    end_date       = "2026-01-01T00:00:00Z"
    interval {
      days    = 1
      seconds = 300
    }
    rotation_start = "2025-01-01T00:00:00Z"
    member {
      user_id = datadog_user.foo.id
    }
    member {
      user_id = null
    }
    member {}
    name = "Primary On-Call Layer"
    restrictions {
      end_day    = "monday"
      end_time   = "17:00:00"
      start_day  = "monday"
      start_time = "09:00:00"
    }
  }
}

