# Manage the evaluation order of all severity modifier rules for the organization.
# rule_ids must list every severity modifier rule ID; rules created outside Terraform appear as drift.

resource "datadog_security_findings_severity_modifier_rules_order" "order" {
  name = "security_findings_severity_modifier_rules_order"
  rule_ids = [
    datadog_security_findings_severity_modifier_rule.escalate_prod_secrets.id,
    # ... add the IDs of every other severity modifier rule in the desired evaluation order.
  ]
}
