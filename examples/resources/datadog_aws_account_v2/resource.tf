# Create new aws_account_v2 resource

resource "datadog_aws_account_v2" "foo" {
  account_tags   = "UPDATE ME"
  aws_account_id = "123456789012"
  aws_partition  = "aws"
  logs_config {
    lambda_forwarder {
      lambdas = "UPDATE ME"
      sources = "UPDATE ME"
    }
  }
  metrics_config {
    automute_enabled          = True
    collect_cloudwatch_alarms = True
    collect_custom_metrics    = True
    enabled                   = True
    tag_filters {
      namespace = "AWS/EC2"
      tags      = "UPDATE ME"
    }
  }
  resources_config {
    cloud_security_posture_management_collection = "UPDATE ME"
    extended_collection                          = "UPDATE ME"
  }
  traces_config {
  }
}