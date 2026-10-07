# This AWS account must already be integrated with Datadog.
data "aws_caller_identity" "current" {}

resource "datadog_aws_wif_intake_mapping" "agent" {
  arn_pattern = "arn:aws:sts::${data.aws_caller_identity.current.account_id}:assumed-role/DatadogAgentRole/*"
}

# Configure the Agent separately with:
# delegated_auth:
#   org_uuid: <YOUR_DATADOG_ORG_UUID>
