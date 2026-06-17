# Create a rule that automatically opens Jira tickets for critical misconfigurations.

resource "datadog_security_findings_ticket_creation_rule" "critical_misconfigs" {
  name    = "Auto-create Jira tickets for critical misconfigurations"
  enabled = true

  rule {
    finding_types = ["misconfiguration"]
    query         = "env:prod @severity:critical"
  }

  action {
    project_id          = "11111111-1111-1111-1111-111111111111"
    target              = "jira"
    assignee_id         = "22222222-2222-2222-2222-222222222222"
    max_tickets_per_day = 50
    # Optional custom Jira fields, JSON-encoded.
    fields = jsonencode({
      labels = ["security"]
    })
  }
}
