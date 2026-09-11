# Create a rule that automatically opens Jira tickets for critical misconfigurations.

resource "datadog_security_findings_ticket_creation_rule" "critical_misconfigs" {
  name    = "Auto-create Jira tickets for critical misconfigurations"
  enabled = true

  rule = {
    finding_types = ["misconfiguration"]
    query         = "env:prod @severity:critical"
  }

  action = {
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

# Create a rule that automatically opens Linear issues for exposed secrets.

resource "datadog_security_findings_ticket_creation_rule" "exposed_secrets" {
  name    = "Auto-create Linear issues for exposed secrets"
  enabled = true

  rule = {
    finding_types = ["secret"]
  }

  action = {
    project_id          = "11111111-1111-1111-1111-111111111111"
    target              = "linear"
    max_tickets_per_day = 25
    # Optional Linear-specific fields, JSON-encoded.
    fields = jsonencode({
      linear_project_id = "33333333-3333-3333-3333-333333333333"
      linear_label_ids  = ["44444444-4444-4444-4444-444444444444"]
    })
  }
}
