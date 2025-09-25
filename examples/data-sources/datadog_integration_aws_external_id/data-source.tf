# Get the external ID for the AWS account "123456789012"
data "datadog_integration_aws_external_id" "example" {
  aws_account_id = "123456789012"
}
