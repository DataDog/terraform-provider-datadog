# Default inbox rules always exist server-side and can only be imported, not created.
# Only `enabled` can be managed; all other attributes are read from Datadog.
# Import the rule with its fixed ID before applying, e.g.:
#   terraform import datadog_security_findings_default_inbox_rule.secret secret_default_rule
# Known default rule IDs: identity_risk_default_rule, secret_default_rule,
# library_vulnerability_default_rule, attack_path_default_rule,
# host_and_container_vulnerability_default_rule, runtime_code_vulnerability_default_rule,
# iac_misconfiguration_default_rule, misconfiguration_default_rule.

resource "datadog_security_findings_default_inbox_rule" "secret" {
  enabled = true
}
