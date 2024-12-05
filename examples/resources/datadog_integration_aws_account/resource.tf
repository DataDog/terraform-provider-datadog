# Create new integration_aws_account resource

resource "datadog_integration_aws_account" "foo" {
  account_tags   = ["env:prod"]
  aws_account_id = "123456789012"
  aws_partition  = "aws"
  logs_config {
    lambda_forwarder {
      lambdas = ["arn:aws:lambda:us-east-1:123456789012:function:my-lambda"]
      sources = ["s3"]
    }
  }
  metrics_config {
    automute_enabled          = true
    collect_cloudwatch_alarms = true
    collect_custom_metrics    = true
    enabled                   = true
    namespace_filters {
      aws_namespace_filters_exclude_only {
        exclude_only = ["AWS/SQS", "AWS/ElasticMapReduce"]
      }
    }
    tag_filters {
      namespace = "AWS/EC2"
      tags      = ["datadog:true"]
    }
  }
  resources_config {
    cloud_security_posture_management_collection = false
    extended_collection                          = true
  }
  traces_config {
    xray_services {
      x_ray_services_include_all {
        include_all = true
      }
    }
  }
}
