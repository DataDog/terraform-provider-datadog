# Manage the evaluation order of all inbox rules for the organization.
# rule_ids must list every inbox rule ID; rules created outside Terraform appear as drift.

resource "datadog_security_findings_inbox_rules_order" "order" {
  name = "security_findings_inbox_rules_order"
  rule_ids = [
    datadog_security_findings_inbox_rule.triage_production_misconfigurations.id,
    # ... add the IDs of every other inbox rule in the desired evaluation order.
  ]
}
