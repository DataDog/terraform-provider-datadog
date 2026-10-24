resource "datadog_feature_flag_environment" "production" {
  name                          = "Production"
  queries                       = ["env:production"]
  is_production                 = true
  require_feature_flag_approval = false
}
