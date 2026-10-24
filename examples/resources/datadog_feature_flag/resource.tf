resource "datadog_feature_flag" "checkout" {
  key                 = "checkout-redesign"
  name                = "Checkout redesign"
  description         = "Enables the redesigned checkout flow"
  value_type          = "BOOLEAN"
  default_variant_key = "off"

  variants = [
    {
      key   = "off"
      name  = "Off"
      value = "false"
    },
    {
      key   = "on"
      name  = "On"
      value = "true"
    },
  ]
}
