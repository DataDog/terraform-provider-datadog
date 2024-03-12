resource "datadog_security_monitoring_suppression" "my_suppression" {
  name        = "example_agent_rule"
  enabled     = true
  description = "im a rule"
  expression  = "open.file.name == \"etc/shadow/password\""
}
