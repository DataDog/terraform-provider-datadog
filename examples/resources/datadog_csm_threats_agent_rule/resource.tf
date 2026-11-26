resource "datadog_csm_threats_agent_rule" "my_agent_rule" {
  name         = "my_agent_rule"
  enabled      = true
  description  = "This is a rule"
  expression   = "open.file.name == \"etc/shadow/password\""
  policy_id    = "jm4-lwh-8cs"
  product_tags = ["compliance_framework:PCI-DSS"]
  actions {
    set {
      name   = "updated_security_actions"
      field  = "exec.file.path"
      append = false
      scope  = "process"
    }
  }
  actions {
    hash {
      field = "open.file"
    }
  }
  actions {
    kill {
      signal = "SIGKILL"
      scope  = "process"
    }
  }
  actions {
    network_filter {
      filter = "tcp port 4444"
      policy = "drop"
      scope  = "cgroup"
    }
  }
}
