data "datadog_integration_aws_external_id" "example" {
  aws_account_id = "123456789012"
  # aws_partition = "aws" # optional, use to disambiguate if needed
}

# Example usage:
# resource "datadog_integration_aws_account" "example" {
#   aws_account_id = data.datadog_integration_aws_external_id.example.aws_account_id
#   aws_partition  = "aws"
#   auth_config {
#     aws_auth_config_role {
#       role_name   = "DatadogIntegrationRole"
#       external_id = data.datadog_integration_aws_external_id.example.external_id
#     }
#   }
# }

