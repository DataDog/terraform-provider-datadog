resource "datadog_role" "example" {
  name = "quota-example"
}

resource "datadog_usage_quota" "example" {
  quota_namespace = "ai_credits"
  scope = {
    role_uuid = datadog_role.example.id
  }
  usage_limit = 1000
  enforced    = true
}
