data "datadog_role" "read_only" {
  filter = "Datadog Read Only Role"
}

resource "datadog_service_account" "terraform" {
  email = "terraform-service-account@example.com"
  name  = "Terraform service account"
  roles = [data.datadog_role.read_only.id]
}

resource "datadog_github_oidc_identity_mapping" "terraform" {
  # A service account's handle is its UUID, so `id` is both stable and already the
  # canonical value Datadog stores. Using `email` here would work, but Datadog
  # normalizes it to the handle, leaving this attribute import-unstable.
  account_identifier = datadog_service_account.terraform.id
  claim_matchers = {
    # The owner in the `sub` prefix must be a literal, not a regex. Replace
    # my-org/terraform with a repository you control: the mapping grants the
    # mapped identity's Datadog access to every workflow that can mint a
    # matching OIDC token, so never configure a repository you do not own.
    sub        = "repo:my-org/terraform:ref:refs/heads/main"
    repository = "^my-org/terraform$"
  }
}
