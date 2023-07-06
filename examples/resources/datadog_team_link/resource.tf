resource "datadog_team" "foo" {
  description = "Example team"
  handle      = "example-team-updated"
  name        = "Example Team-updated"
}

# Create new team_link resource

resource "datadog_team_link" "foo" {
  team_id  = datadog_team.foo.id
  label    = "Link label"
  position = "Example link"
  url      = "https://example.com"
}