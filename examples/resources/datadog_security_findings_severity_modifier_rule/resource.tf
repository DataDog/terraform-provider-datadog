# Create a new severity modifier rule that raises the severity of secrets found in production.

resource "datadog_security_findings_severity_modifier_rule" "escalate_prod_secrets" {
  name    = "Escalate secrets found in production"
  enabled = true

  rule = {
    finding_types = ["secret"]
    query         = "env:prod"
  }

  action = {
    set = {
      severity    = "critical"
      description = "Secrets in production are always treated as critical."
    }
  }
}

# Create a severity modifier rule that shifts the severity of a finding down by one rank.

resource "datadog_security_findings_severity_modifier_rule" "deprioritize_dev_misconfigurations" {
  name    = "Deprioritize misconfigurations in dev"
  enabled = true

  rule = {
    finding_types = ["misconfiguration"]
    query         = "env:dev"
  }

  action = {
    shift = {
      severity_delta = "down_one"
      description    = "Misconfigurations in dev are lower priority."
    }
  }
}
