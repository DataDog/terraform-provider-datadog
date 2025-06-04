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
    escalation_policy = "00000000-aba2-0000-0000-000000000000"
    urgency           = "dynamic"
  }
}
