resource "datadog_action_connection" "aws" {
  name = var.connection_name

  aws {
    assume_role {
      account_id = var.aws_account_id
      role       = var.aws_role_name
    }
  }
}
