resource "datadog_csm_threats_agent_rule" "my_agent_rule" {
  name        = "my_agent_rule"
  enabled     = true
  description = "im a rule"
  expression  = "open.file.name == \"etc/shadow/password\""
}