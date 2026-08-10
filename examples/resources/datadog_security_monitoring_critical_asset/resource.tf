resource "datadog_security_monitoring_critical_asset" "my_critical_asset" {
  description = "Production database servers that should trigger critical alerts"
  enabled     = true
  query       = "env:production service:database"
  rule_query  = "type:(log_detection OR signal_correlation OR workload_security OR application_security) ruleId:007-d1a-1f3"
  severity    = "increase"
  tags        = ["env:production", "team:security"]
}
