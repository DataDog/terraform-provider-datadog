# Create new team-hierarchy-links resource

resource "datadog_team-hierarchy-links" "foo" {
    body {
    data {
    relationships {
    parent_team {
    data {
    id = "692e8073-12c4-4c71-8408-5090bd44c9c8"
    type = "team"
    }
    }
    sub_team {
    data {
    id = "692e8073-12c4-4c71-8408-5090bd44c9c8"
    type = "team"
    }
    }
    }
    type = "team_hierarchy_links"
    }
    }
}
