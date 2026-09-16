# Assign ownership of checkout views for a browser RUM application.

resource "datadog_rum_team_ownership" "checkout" {
  application_id = "<APPLICATION_ID>"
  team_handle    = "web-platform"
  view_name      = "/checkout"
  match_type     = "prefix"
}
