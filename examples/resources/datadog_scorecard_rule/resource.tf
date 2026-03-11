resource "datadog_scorecard_rule" "example" {
  name           = "Has a README"
  scorecard_name = "Service Maturity"
  description    = "Checks if the service has a README file."
  enabled        = true
  level          = "1"
  owner          = "my-team"
}
