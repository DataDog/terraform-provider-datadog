resource "datadog_feature_flag" "checkout" {
  key                 = "checkout-redesign"
  name                = "Checkout redesign"
  value_type          = "BOOLEAN"
  default_variant_key = "off"
  variants = [
    { key = "off", name = "Off", value = "false" },
    { key = "on", name = "On", value = "true" },
  ]
}

resource "datadog_feature_flag_environment" "production" {
  name    = "Production"
  queries = ["env:production"]
}

resource "datadog_feature_flag_environment_config" "checkout_production" {
  feature_flag_id = datadog_feature_flag.checkout.id
  environment_id  = datadog_feature_flag_environment.production.id
  enabled         = true
}
