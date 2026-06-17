# Manage the evaluation order of all due date rules for the organization.
# This resource owns the full ordering: any due date rule you omit here (for example
# one created from the UI) is appended to the end of the order and reported as a
# warning on apply. List every rule ID to control their exact position.

resource "datadog_security_findings_due_date_rules_order" "order" {
  name = "security_findings_due_date_rules_order"
  rule_ids = [
    datadog_security_findings_due_date_rule.prod_sla.id,
    # ... add the IDs of every other due date rule in the desired evaluation order.
  ]
}
