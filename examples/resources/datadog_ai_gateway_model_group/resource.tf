resource "datadog_ai_gateway_model_group" "coding" {
  name                = "Coding models"
  model_ids           = [101, 102]
  provider_ids        = [3]
  routing_rule_slugs  = ["coding-route"]
  allow_fast_mode     = false
  max_reasoning_level = "high"
}
