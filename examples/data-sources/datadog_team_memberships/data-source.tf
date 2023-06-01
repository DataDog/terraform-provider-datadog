data "datadog_team_memberships" "foo" {
  team_id        = "e6723c40-edb1-11ed-b816-da7ad0900002"
  filter_keyword = "foo@example.com"
}
