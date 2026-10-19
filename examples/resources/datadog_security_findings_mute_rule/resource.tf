# Create a new mute rule that suppresses accepted risks in dev environments.

resource "datadog_security_findings_mute_rule" "accepted_risks_dev" {
  name    = "Mute accepted risks in dev"
  enabled = true

  rule = {
    finding_types = ["misconfiguration"]
    query         = "env:dev team:platform @severity:low"
  }

  action = {
    reason             = "risk_accepted"
    reason_description = "Accepted for dev environments only"
    # Optional Unix timestamp in milliseconds at which the mute expires.
    expire_at = 4070908800000
  }
}
