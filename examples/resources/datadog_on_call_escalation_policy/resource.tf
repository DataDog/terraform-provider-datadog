resource "datadog_on_call_escalation_policy" "policy_test" {
  name                       = "Policy Name"
  resolve_page_on_policy_end = true
  retries                    = 3
  step {
    assignment             = "round-robin"
    escalate_after_seconds = 300
    target {
      team = "00000000-aba2-0000-0000-000000000000"
    }
    target {
      user = "00000000-aba2-0000-0000-000000000000"
    }
    target {
      schedule = "00000000-aba2-0000-0000-000000000000"
    }
  }
}
