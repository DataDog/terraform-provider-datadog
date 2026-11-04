variable "agent_aws_account_id" {
  type        = string
  description = "The AWS account where the Agent runs. It must be integrated with Datadog and may differ from the account running Terraform."
}

# If this configuration also manages the AWS integration, replace
# var.agent_aws_account_id below with datadog_integration_aws_account.agent.aws_account_id
# (using your integration resource's name) to establish creation order.
resource "datadog_aws_wif_intake_mapping" "agent" {
  arn_pattern = "arn:aws:sts::${var.agent_aws_account_id}:assumed-role/DatadogAgentRole/*"
}

# Configure the Agent separately with:
# delegated_auth:
#   org_uuid: <YOUR_DATADOG_ORG_UUID>
