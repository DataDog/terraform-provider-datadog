resource "datadog_deployment_gate" "all_environments" {
  service    = "my-service"
  identifier = "my-gate"
  env        = "*"

  rule {
    name = "monitor"
    type = "monitor"
    options {
      query    = "service:my-service"
      duration = 300
    }
  }
}
