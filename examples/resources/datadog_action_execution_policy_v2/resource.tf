resource "datadog_action_execution_policy_v2" "example" {
  name   = "Deny maintenance scripts"
  effect = "deny"
  action_pattern = {
    integration = "INTEGRATION_SCRIPT"
    action_fqns = ["*"]
  }
  scope = {
    scripts = {
      rules = [{ target_script_names = ["maintenance"] }]
    }
  }
}
