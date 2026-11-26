# Warning: the kill action terminates every process that matches the expression.
# This example uses a dedicated test file and starts disabled. Review the expression before you enable it.
resource "datadog_csm_threats_agent_rule" "my_agent_rule" {
  name         = "my_agent_rule"
  enabled      = false
  description  = "This is a rule"
  expression   = "open.file.path == \"/tmp/terraform-kill-test\""
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
