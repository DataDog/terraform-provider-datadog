resource "datadog_security_notification_rule" "signal_rule" {
  name = "My signal notification rule"
  selectors {
    trigger_source = "security_signals"
    rule_types     = ["workload_security"]
    query          = "env:prod"
  }
  enabled = false
  targets = ["bob@email.com", "alice@email.com"]
}

resource "datadog_security_notification_rule" "vulnerability_rule" {
  name = "My vulnerability notification rule"
  selectors {
    trigger_source = "security_findings"
    rule_types     = ["application_library_vulnerability", "identity_risk"]
    severities     = ["critical", "high"]
  }
  time_aggregation = 36000
  targets          = ["john@email.com"]
}
