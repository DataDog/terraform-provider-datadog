data "datadog_scorecard_rules" "enabled_rules" {
  filter_enabled = true
  filter_name    = "Has a README"
}
