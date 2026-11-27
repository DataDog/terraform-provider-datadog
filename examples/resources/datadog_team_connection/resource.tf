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

# Create a connection between a Datadog team and a ServiceNow group
resource "datadog_team_connection" "servicenow" {
  team {
    id   = datadog_team.example.id
    type = "team"
  }
  connected_team {
    id   = "my-instance/0123456789abcdef0123456789abcdef"
    type = "servicenow_team"
  }
  source = "servicenow"
}
