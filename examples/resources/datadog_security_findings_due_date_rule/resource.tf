# Create a due date rule that assigns remediation SLAs to production misconfigurations.

resource "datadog_security_findings_due_date_rule" "prod_sla" {
  name    = "SLA for production misconfigurations"
  enabled = true

  rule {
    finding_types = ["misconfiguration"]
    query         = "env:prod"
  }

  action {
    due_from = "first_seen"

    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
    due_days_per_severity {
      severity    = "high"
      due_in_days = 30
    }
    due_days_per_severity {
      severity    = "medium"
      due_in_days = 90
    }
  }
}
