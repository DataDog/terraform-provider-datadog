# Manage the evaluation order of all mute rules for the organization.
# This resource owns the full ordering: any mute rule you omit here (for example
# one created from the UI) is appended to the end of the order and reported as a
# warning on apply. List every rule ID to control their exact position.

resource "datadog_security_findings_mute_rules_order" "order" {
  name = "security_findings_mute_rules_order"
  rule_ids = [
    datadog_security_findings_mute_rule.accepted_risks_dev.id,
    # ... add the IDs of every other mute rule in the desired evaluation order.
  ]
}
