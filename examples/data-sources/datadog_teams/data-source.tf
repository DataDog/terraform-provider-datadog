data "datadog_teams" "example" {
  filter_keyword = "team-member@company.com"
  filter_me      = true
}
