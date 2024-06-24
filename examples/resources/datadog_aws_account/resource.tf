# Create new aws_account resource

resource "datadog_aws_account" "foo" {
  account_tags = "UPDATE ME"
  auth_config {
    access_key_id     = "UPDATE ME"
    external_id       = "UPDATE ME"
    role_name         = "UPDATE ME"
    secret_access_key = "UPDATE ME"
  }
  aws_account_id = "123456789012"
  aws_regions {
    include_all  = "UPDATE ME"
    include_only = "UPDATE ME"
  }
  logs_config {
    lambda_forwarder {
      lambdas = "UPDATE ME"
      sources = "UPDATE ME"
    }
  }
  metrics_config {
    automute_enabled          = "UPDATE ME"
    collect_cloudwatch_alarms = "UPDATE ME"
    collect_custom_metrics    = "UPDATE ME"
    enabled                   = "UPDATE ME"
    namespace_filters {
      exclude_all  = "UPDATE ME"
      exclude_only = "UPDATE ME"
      include_all  = "UPDATE ME"
      include_only = "UPDATE ME"
    }
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
    xray_services {
      include_all  = "UPDATE ME"
      include_only = "UPDATE ME"
    }
  }
}