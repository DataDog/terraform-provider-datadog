resource "datadog_team" "foo" {
  description = "Example team"
  handle      = "example-team-updated"
  name        = "Example Team-updated"
}

resource "datadog_team_permission_setting" "foo" {
  team_id = datadog_team.foo.id
  action  = "manage_membership"
  value   = "organization"
}