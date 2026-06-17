# Manage the evaluation order of all ticket creation rules for the organization.
# This resource owns the full ordering: any ticket creation rule you omit here (for
# example one created from the UI) is appended to the end of the order and reported as
# a warning on apply. List every rule ID to control their exact position.

resource "datadog_security_findings_ticket_creation_rules_order" "order" {
  name = "security_findings_ticket_creation_rules_order"
  rule_ids = [
    datadog_security_findings_ticket_creation_rule.critical_misconfigs.id,
    # ... add the IDs of every other ticket creation rule in the desired evaluation order.
  ]
}
