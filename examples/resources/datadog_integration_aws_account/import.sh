# AWS Account Config ID can be retrieved by using the List all AWS integrations endpoint and querying by AWS Account ID:
# https://docs.datadoghq.com/api/latest/aws-integration/#list-all-aws-integrations


terraform import datadog_integration_aws_account.example "<datadog-aws-account-config-id>"
