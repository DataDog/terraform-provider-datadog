# AWS Account Config ID can be retrieved by using the List all AWS integrations endpoint and querying by AWS Account ID.
# https://docs.datadoghq.com/api/latest/aws-integration/#list-all-aws-integrations
#
# The `datadog_integration_aws_account` resource encompasses configuration that was previously managed under
# these separate resources:
#    - datadog_integration_aws
#    - datadog_integration_aws_lambda_arn
#    - datadog_integration_aws_log_collection
#    - datadog_integration_aws_tag_filter
#
# To migrate your account configuration from `datadog_integration_aws*` resources to `datadog_integration_aws_account`:
# 1. Import your integrated accounts into `datadog_integration_aws_account` resources using the config below.
# 2. Run `terraform state rm` to delete all resources of the following types from state:
#    - datadog_integration_aws
#    - datadog_integration_aws_lambda_arn
#    - datadog_integration_aws_log_collection
#    - datadog_integration_aws_tag_filter
#


terraform import datadog_integration_aws_account.example "<datadog-aws-account-config-id>"
