resource "datadog_security_monitoring_suppression" "my_suppression" {
  name                 = "My suppression"
  description          = "Suppression for low severity CloudTrail signals from John Doe, excluding test environments from analysis, limited to 2024"
  enabled              = true
  rule_query           = "severity:low source:cloudtrail"
  suppression_query    = "@usr.id:john.doe"
  data_exclusion_query = "env:test"
  start_date           = "2024-12-01T16:00:00Z"
  expiration_date      = "2024-12-31T12:00:00Z"
}
