# Create new team-hierarchy-links resource

resource "datadog_team-hierarchy-links" "foo" {
  parent_team_id = "692e8073-12c4-4c71-8408-5090bd44c9c8"
  sub_team_id    = "d10bf972-ca7c-4181-9668-6db783533f6e"
}
