terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
    }
  }
}

variable "datadog_api_key" {
  description = "Datadog API Key"
}

variable "datadog_app_key" {
  description = "Datadog App Key"
}

provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
  api_url = "https://app.datad0g.com"
}

# First create a policy
resource "datadog_csm_threats_policy" "test_policy" {
  name            = "test_policy_sarra_v2"
  enabled         = true
  description     = "Test policy for agent rules"
  host_tags_lists = [[]]
}

resource "datadog_csm_threats_agent_rule" "test" {
  policy_id   = datadog_csm_threats_policy.test_policy.id
  name        = "test_rule_v"
  description = "Test rule with different optional fields"
  enabled     = true
  expression  = "open.file.name == \"passwd\""
  actions     = [
    {
      set = {
        name   = "updated_security_action_v2"
        field  = ""
        value  = ""
        append = true
        scope  = ""
        ttl    = 600
        size   = 0
      }
    }
  ]
} 