data "aws_caller_identity" "current" {}

data "datadog_role" "read_only" {
  filter = "Datadog Read Only Role"
}

resource "datadog_service_account" "terraform" {
  email = "terraform-service-account@example.com"
  name  = "Terraform service account"
  roles = [data.datadog_role.read_only.id]
}

resource "datadog_aws_wif_persona_mapping" "terraform" {
  # A service account's handle is its UUID, so `id` is both stable and already the
  # canonical value Datadog stores. Using `email` here would work, but Datadog
  # normalizes it to the handle, leaving this attribute import-unstable.
  account_identifier = datadog_service_account.terraform.id
  arn_pattern        = "arn:aws:sts::${data.aws_caller_identity.current.account_id}:assumed-role/terraform-runner/*"
}
