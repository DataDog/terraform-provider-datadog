# Create new aws_account_v2 resource

resource "datadog_aws_account_v2" "foo" {
  account_tags   = []
  aws_account_id = "123456789012"
  aws_partition  = "aws"
  logs_config {
    lambda_forwarder {
      lambdas = []
      sources = []
    }
  }
  metrics_config {
    automute_enabled          = true
    collect_cloudwatch_alarms = true
    collect_custom_metrics    = true
    enabled                   = true
    tag_filters {
      namespace = "AWS/EC2"
      tags      = ["test:test"]
    }
  }
  resources_config {
    cloud_security_posture_management_collection = true
    extended_collection                          = true
  }
  traces_config {}
}
