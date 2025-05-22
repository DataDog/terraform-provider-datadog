terraform {
 required_providers {
   datadog = {
     source = "DataDog/datadog"
   }
 }
}

resource "datadog_user" "foo" {
    email = "USER_EMAIL"
}

resource "datadog_team" "foo" {
    description = "Description"
    handle      = "TEAM_HANDLE"
    name        = "TEAM_NAME"
}

resource "datadog_on_call_schedule" "schedule" {
  name      = "Escalation Policy Test Schedule UNIQ"
  time_zone = "America/New_York"
  teams     = [datadog_team.foo.id]
  layer {
    effective_date = "2025-01-01T00:00:00Z"
    interval {
      days    = 1
      seconds = 300
    }
    rotation_start = "2025-01-01T00:00:00Z"
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



resource "datadog_on_call_escalation_policy" "policy" {
    name                        = "POLICY_NAME"
    resolve_page_on_policy_end = true
    retries = 3
    step {
        assignment = "round-robin"
        escalate_after_seconds = 300
        target {
            team = datadog_team.foo.id
        }
        target {
            user = datadog_user.foo.id
        }
        target {
            schedule = datadog_on_call_schedule.schedule.id
        }
    }
}

