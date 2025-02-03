# Create new integration_aws_account resource

resource "datadog_integration_aws_account" "foo" {
  account_tags   = ["env:prod"]
  aws_account_id = "123456789012"
  aws_partition  = "aws"
  aws_regions {
    include_all = true
  }

  auth_config {
    aws_auth_config_role {
      role_name = "DatadogIntegrationRole"
    }
  }
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
      exclude_only = ["AWS/SQS", "AWS/ElasticMapReduce"]
    }
    tag_filters {
      namespace = "AWS/EC2"
      tags      = ["datadog:true"]
    }
  }
  resources_config {
    cloud_security_posture_management_collection = true
    extended_collection                          = true
  }
  traces_config {
    xray_services {
      include_all = true
    }
  }
}

# Create new integration_aws_account resource with all Datadog-provided defaults configured
resource "datadog_integration_aws_account" "foo-defaults" {
  aws_account_id = "234567890123"
  aws_partition  = "aws"
  aws_regions {}

  auth_config {
    aws_auth_config_role {
      role_name = "DatadogIntegrationRole"
    }
  }
  logs_config {
    lambda_forwarder {}
  }
  metrics_config {
    namespace_filters {}
  }
  resources_config {}
  traces_config {
    xray_services {}
  }
}
