# Manage the evaluation order of all ticket creation rules for the organization.
# rule_ids must list every ticket creation rule ID; rules created outside Terraform appear as drift.

resource "datadog_security_findings_ticket_creation_rules_order" "order" {
  name = "security_findings_ticket_creation_rules_order"
  rule_ids = [
    datadog_security_findings_ticket_creation_rule.critical_misconfigs.id,
    # ... add the IDs of every other ticket creation rule in the desired evaluation order.
  ]
}
