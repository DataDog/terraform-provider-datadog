
resource "datadog_team" "reroute_source" {
  description = "Source team for reroute_to_team test"
  handle      = "TEAM_HANDLE-src"
  name        = "TEAM_NAME-src"
}

resource "datadog_team" "reroute_dest_a" {
  description = "Destination team A for reroute_to_team test"
  handle      = "TEAM_HANDLE-dst-a"
  name        = "TEAM_NAME-dst-a"
}

resource "datadog_team" "reroute_dest_b" {
  description = "Destination team B for reroute_to_team test"
  handle      = "TEAM_HANDLE-dst-b"
  name        = "TEAM_NAME-dst-b"
}

resource "datadog_on_call_team_routing_rules" "reroute_test" {
  id = datadog_team.reroute_source.id
  rule {
    query = "priority:2"
    action {
      reroute_to_team {
        destination_team_id = datadog_team.FIRST_DEST.id
      }
    }
  }

  rule {
    action {
      reroute_to_team {
        destination_team_id = datadog_team.SECOND_DEST.id
      }
    }
  }
}
