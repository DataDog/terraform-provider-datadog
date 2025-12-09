
resource "datadog_user" "policy_test" {
  email = "USER_EMAIL"
}

resource "datadog_team" "policy_test" {
  description = "Description"
  handle      = "TEAM_HANDLE"
  name        = "TEAM_NAME"
}

resource "datadog_on_call_schedule" "schedule" {
  name      = "Escalation Policy Test Schedule UNIQ"
  time_zone = "America/New_York"
  teams     = [datadog_team.policy_test.id]
  layer {
    effective_date = "2025-01-01T00:00:00Z"
    interval {
      days    = 1
      seconds = 300
    }
    rotation_start = "2025-01-01T00:00:00Z"
    users          = [datadog_user.policy_test.id, null]
    name           = "Primary On-Call Layer"
    restriction {
      end_day    = "monday"
      end_time   = "17:00:00"
      start_day  = "monday"
      start_time = "09:00:00"
    }
  }
}

resource "datadog_on_call_escalation_policy" "policy_test" {
  name                       = "POLICY_NAME"
  resolve_page_on_policy_end = true
  step {
    assignment             = "round-robin"
    escalate_after_seconds = 300
    target {
      team = datadog_team.policy_test.id
    }
    target {
      user = datadog_user.policy_test.id
    }
    target {
      schedule = datadog_on_call_schedule.schedule.id
    }
  }

  step {
    escalate_after_seconds = 600
    target {
      schedule          = datadog_on_call_schedule.schedule.id
      schedule_position = "next"
    }
  }
}

# This policy is used to test the defaults of the escalation policy resource.
# ie missing fields are set to their default values.
resource "datadog_on_call_escalation_policy" "policy_test_defaults" {
  name = "policy_test_defaults"
  step {
    escalate_after_seconds = 100
    target {
      schedule = datadog_on_call_schedule.schedule.id
    }
  }
}
