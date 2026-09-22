# Create a new inbox rule that pushes prod misconfigurations into the Security Inbox.

resource "datadog_security_findings_inbox_rule" "triage_production_misconfigurations" {
  name    = "Triage production misconfigurations"
  enabled = true

  rule = {
    finding_types = ["misconfiguration"]
    query         = "env:prod team:platform"
  }

  action = {
    # An optional description providing more context for the rule.
    description = "Needs triage"
  }
}
