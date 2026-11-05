# Kubernetes: allow read-only actions against prod namespaces, on agents/PARs
# matching either target tag set below.
resource "datadog_action_execution_policy" "kubernetes_prod_read" {
  name   = "allow-kubernetes-prod-read"
  effect = "allow"

  action_pattern {
    integration = "INTEGRATION_KUBERNETES"
    action_fqns = ["com.datadoghq.kubernetes.core.getPod", "com.datadoghq.kubernetes.core.listPod"]
  }

  scope {
    kubernetes {
      rule {
        target_namespaces = ["prod", "prod-system"]
      }
    }
  }

  target {
    name       = "prod-us"
    agent_tags = ["env:prod", "region:us-east-1"]
  }

  target {
    name       = "prod-eu"
    agent_tags = ["env:prod", "region:eu-west-1"]
  }
}

# Remote action (rshell): allow command execution restricted to a read-only path on staging agents.
resource "datadog_action_execution_policy" "rshell_staging_readonly" {
  name   = "rshell-staging-readonly"
  effect = "allow"

  action_pattern {
    integration = "INTEGRATION_REMOTE_ACTION"
    action_fqns = ["com.datadoghq.remoteaction.rshell.runCommand"]
  }

  scope {
    remote_action_rshell {
      rule {
        target_paths = ["/etc/datadog-agent/"]
        access       = "read_only"
      }
    }
  }

  target {
    name       = "staging"
    agent_tags = ["env:staging"]
  }
}

# Scripts: deny a specific shell script action everywhere, regardless of agent tags
# (no `target` block means the policy applies fleet-wide).
resource "datadog_action_execution_policy" "deny_dangerous_script" {
  name   = "deny-dangerous-cleanup-script"
  effect = "deny"

  action_pattern {
    integration = "INTEGRATION_SCRIPT"
    action_fqns = ["com.datadoghq.script.runShellScript"]
  }

  scope {
    scripts {
      rule {
        target_script_names = ["dangerous-cleanup.sh"]
      }
    }
  }
}

# Access to a policy is managed separately via datadog_restriction_policy,
# keyed by "execution-policy:<id>".
resource "datadog_restriction_policy" "kubernetes_prod_perms" {
  resource_id = "execution-policy:${datadog_action_execution_policy.kubernetes_prod_read.id}"

  bindings {
    relation   = "editor"
    principals = ["role:sre-oncall"]
  }
}
