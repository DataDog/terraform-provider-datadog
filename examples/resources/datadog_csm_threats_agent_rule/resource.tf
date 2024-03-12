resource "datadog_csm_threats_agent_rule" "my_agent_rule" {
  name        = "my_agent_rule"
  description = "My agent rule"
  enabled     = true
  expression  = "exec.file.name == \"java\""
}