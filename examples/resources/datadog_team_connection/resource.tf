resource "datadog_team" "example" {
  description = "Example team"
  handle      = "example-team"
  name        = "Example Team"
}

# Create a connection between a Datadog team and a GitHub team
resource "datadog_team_connection" "example" {
  team {
    id   = datadog_team.example.id
    type = "team"
  }
  connected_team {
    id   = "@GitHubOrg/team-handle"
    type = "github_team"
  }
}
