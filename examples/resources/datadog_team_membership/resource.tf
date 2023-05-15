resource "datadog_team" "foo" {
  description = "Example team"
  handle      = "example-team-updated"
  name        = "Example Team-updated"
}

resource "datadog_user" "foo" {
  email = "new@example.com"
}

# Create new team_membership resource
resource "datadog_team_membership" "foo" {
  team_id = datadog_team.foo.id
  user_id = datadog_user.foo.id
  role    = "admin"
}