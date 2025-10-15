# Create new deployment_rule resource

resource "datadog_deployment_rule" "foo" {
  gate_id = "UPDATE ME"
  dry_run = "UPDATE ME"
  name    = "My deployment rule"
  type    = "faulty_deployment_detection"
}