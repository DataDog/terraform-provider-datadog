
resource "datadog_team" "stable_ids_test" {
  description = "Test team for stable rule IDs"
  handle      = "TEAM_HANDLE"
  name        = "TEAM_NAME"
}

resource "datadog_on_call_escalation_policy" "stable_ids_test" {
  name                       = "POLICY_NAME"
  resolve_page_on_policy_end = true
  retries                    = 3
  step {
    assignment             = "round-robin"
    escalate_after_seconds = 300
    target {
      team = datadog_team.stable_ids_test.id
    }
  }
}

resource "datadog_on_call_team_routing_rules" "stable_ids_test" {
  id = datadog_team.stable_ids_test.id
  rule {
    query             = "QUERY_1"
    urgency           = "low"
    escalation_policy = datadog_on_call_escalation_policy.stable_ids_test.id
  }
  rule {
    urgency           = "dynamic"
    escalation_policy = datadog_on_call_escalation_policy.stable_ids_test.id
  }
}
