# Create new team-notification-rule resource

resource "datadog_team" "foo" {
  description = "Example team"
  handle      = "example-team"
  name        = "Example Team"
}

resource "datadog_team_notification_rule" "foo" {
  team_id = datadog_team.foo.id
  email {
    enabled = true
  }
  ms_teams {
    connector_name = "test-teams-handle"
  }
  pagerduty {
    service_name = "my-service"
  }
  slack {
    channel   = "#test-channel"
    workspace = "Datadog"
  }
}
