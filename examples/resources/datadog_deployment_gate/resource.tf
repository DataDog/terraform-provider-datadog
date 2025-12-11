# Create new deployment_gate resource

resource "datadog_deployment_gate" "foo" {
  dry_run    = "false"
  env        = "production"
  identifier = "my-gate"
  service    = "my-service"

  rule {
    name    = "fdd"
    type    = "faulty_deployment_detection"
    dry_run = false
    options {
      duration           = 1300
      excluded_resources = ["GET api/v1/test"]
    }
  }

  rule {
    name    = "monitor"
    type    = "monitor"
    dry_run = false
    options {
      query    = "service:test-service"
      duration = 1300
    }
  }
}