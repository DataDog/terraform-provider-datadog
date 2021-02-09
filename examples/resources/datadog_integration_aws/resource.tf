# Create a new Datadog - Amazon Web Services integration
resource "datadog_integration_aws" "sandbox" {
  account_id  = "1234567890"
  role_name   = "DatadogAWSIntegrationRole"
  filter_tags = ["key:value"]
  host_tags   = ["key:value", "key2:value2"]
  account_specific_namespace_rules = {
    auto_scaling = false
    opsworks     = false
  }
  excluded_regions = ["us-east-1", "us-west-2"]
}
