# Get the Datadog AWS integration config ID for the AWS account "123456789012"
data "datadog_integration_aws_account" "example" {
  aws_account_id = "123456789012"
}

# Use it to import the existing integration into the datadog_integration_aws_account resource
import {
  to = datadog_integration_aws_account.example
  id = data.datadog_integration_aws_account.example.id
}
