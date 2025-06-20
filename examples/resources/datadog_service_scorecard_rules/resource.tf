resource "datadog_service_scorecard_rule" "foo" {
  name           = "my-scorecard-rule"
  description    = "My Custom Scorecard Rule"
  enabled        = true
  owner          = "platform"
  scorecard_name = "my_scorecard"
  level          = 2
  scope_query    = "env:prod team:platform"
}