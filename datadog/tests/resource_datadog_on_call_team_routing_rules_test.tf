
resource "datadog_user" "team_rules_test" {
  email = "USER_EMAIL"
}

resource "datadog_team" "team_rules_test" {
  description = "Description"
  handle      = "TEAM_HANDLE"
  name        = "TEAM_NAME"
}

resource "datadog_on_call_schedule" "team_rules_test" {
  name      = "Escalation Policy Test Schedule UNIQ"
  time_zone = "America/New_York"
  teams     = [datadog_team.team_rules_test.id]
  layer {
    effective_date = "2025-01-01T00:00:00Z"
    interval {
      days    = 1
      seconds = 300
    }
    rotation_start = "2025-01-01T00:00:00Z"
    users          = [datadog_user.team_rules_test.id, null]
    name           = "Primary On-Call Layer"
    restriction {
      end_day    = "monday"
      end_time   = "17:00:00"
      start_day  = "monday"
      start_time = "09:00:00"
    }
  }
}

resource "datadog_workflow_automation" "team_rules_test" {
  name        = "Team Rules Test Workflow UNIQ"
  description = "Workflow used by on-call team routing rules test."
  tags        = ["service:foo", "team:bar"]
  published   = true
  spec_json = jsonencode({
    handle = "team-rules-test-workflow-UNIQ"
    steps  = []
    triggers = [
      {
        startStepNames = []
        onCallTrigger  = {}
      }
    ]
  })
}

resource "datadog_on_call_escalation_policy" "team_rules_test" {
  name                        = "POLICY_NAME"
  resolve_page_on_policy_end = true
  retries = 3
  step {
    assignment = "round-robin"
    escalate_after_seconds = 300
    target {
      team = datadog_team.team_rules_test.id
    }
    target {
      user = datadog_user.team_rules_test.id
    }
    target {
      schedule = datadog_on_call_schedule.team_rules_test.id
    }
  }
}

resource "datadog_on_call_team_routing_rules" "team_rules_test" {
  id         = datadog_team.team_rules_test.id
  depends_on = [datadog_workflow_automation.team_rules_test]
  rule {
    query = "tags.service:test"
    action {
      trigger_workflow_automation {
        handle = "team-rules-test-workflow-UNIQ"
      }
    }
    time_restrictions {
      time_zone = "America/New_York"
      restriction {
        end_day    = "monday"
        end_time   = "17:00:00"
        start_day  = "monday"
        start_time = "09:00:00"
      }
    }
  }

  rule {
    query = "tags.service:payment"
    action {
      escalation_policy {
        policy_id = datadog_on_call_escalation_policy.team_rules_test.id
        ack_timeout_minutes = 30
        urgency = "low"
        support_hours {
          time_zone = "America/New_York"
          restriction {
            start_day = "monday"
            start_time = "09:00:00"
            end_day = "friday"
            end_time = "17:00:00"
          }
          restriction {
            start_day  = "tuesday"
            start_time = "09:00:00"
            end_day    = "tuesday"
            end_time   = "17:00:00"
          }
          restriction {
            start_day  = "wednesday"
            start_time = "09:00:00"
            end_day    = "wednesday"
            end_time   = "17:00:00"
          }
          restriction {
            start_day  = "thursday"
            start_time = "09:00:00"
            end_day    = "thursday"
            end_time   = "17:00:00"
          }
          restriction {
            start_day  = "friday"
            start_time = "09:00:00"
            end_day    = "friday"
            end_time   = "17:00:00"
          }
        }
      }
    }
  }

  rule {
    escalation_policy = datadog_on_call_escalation_policy.team_rules_test.id
    urgency = "dynamic"
  }
}
