resource "datadog_on_call_team_routing_rules" "team_rules_test" {
  id = "00000000-aba2-0000-0000-000000000000"
  rule {
    query = "tags.service:test"
    action {
      send_slack_message {
        workspace = "workspace"
        channel   = "channel"
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
        policy_id           = "00000000-aba2-0000-0000-000000000000"
        ack_timeout_minutes = 30
        urgency             = "low"
        support_hours {
          time_zone = "America/New_York"
          restriction {
            start_day  = "monday"
            start_time = "09:00:00"
            end_day    = "monday"
            end_time   = "17:00:00"
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

  # The last rule must be a catch-all: no query, no time restriction,
  # and an escalation policy.
  rule {
    escalation_policy = "00000000-aba2-0000-0000-000000000000"
    urgency           = "dynamic"
  }
}
