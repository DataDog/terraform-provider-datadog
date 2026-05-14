# Configure team synchronization from GitHub
resource "datadog_team_sync" "example" {
  source          = "github"
  type            = "link"
  frequency       = "continuously"
  sync_membership = true

  selection_state {
    operation = "include"
    scope     = "subtree"
    external_id {
      type  = "organization"
      value = "12345"
    }
  }
}
