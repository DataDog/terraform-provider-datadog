# Create new deployment_gate resource

resource "datadog_deployment_gate" "foo" {
  dry_run    = "UPDATE ME"
  env        = "production"
  identifier = "my-gate"
  service    = "my-service"
}