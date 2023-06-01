# Create new team resource

resource "datadog_team" "foo" {
  description = "Team description"
  handle      = "example-team"
  name        = "Example Team"
}