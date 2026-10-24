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

resource "datadog_feature_flag_allocations" "checkout_production" {
  feature_flag_id = datadog_feature_flag.checkout.id
  environment_id  = datadog_feature_flag_environment.production.id
  depends_on      = [datadog_feature_flag_environment_config.checkout_production]

  allocations = [
    {
      key  = "checkout-redesign-staff-production"
      name = "Checkout redesign for staff"
      type = "FEATURE_GATE"

      targeting_rules = [
        {
          conditions = [
            {
              attribute = "user.tier"
              operator  = "EQUALS"
              value     = ["staff"]
            },
          ]
        },
      ]

      variant_weights = [
        {
          variant_key = "on"
          value       = 100
        },
      ]
    },
  ]
}
