resource "datadog_security_monitoring_suppression" "my_suppression" {
  name              = "My suppression"
  description       = "Suppression for low severity CloudTrail signals from test environments limited to 2024"
  enabled           = true
  rule_query        = "severity:low source:cloudtrail"
  suppression_query = "env:test"
  expiration_date   = "2024-12-31T12:00:00Z"
}
