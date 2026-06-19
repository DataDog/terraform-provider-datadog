# Manage the evaluation order of all mute rules for the organization.
# rule_ids must list every mute rule ID; rules created outside Terraform appear as drift.

resource "datadog_security_findings_mute_rules_order" "order" {
  name = "security_findings_mute_rules_order"
  rule_ids = [
    datadog_security_findings_mute_rule.accepted_risks_dev.id,
    # ... add the IDs of every other mute rule in the desired evaluation order.
  ]
}
