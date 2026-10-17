resource "datadog_github_oidc_intake_mapping" "terraform" {
  claim_matchers = {
    # The owner in the `sub` prefix must be a literal, not a regex. Replace
    # my-org/terraform with a repository you control: the mapping grants intake
    # access to every workflow that can mint a matching OIDC token, so never
    # configure a repository you do not own.
    sub        = "repo:my-org/terraform:ref:refs/heads/main"
    repository = "^my-org/terraform$"
  }
}
