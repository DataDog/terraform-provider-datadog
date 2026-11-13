# Spread coding traffic 70/30 between two models on the production provider, falling back to another routing rule.
resource "datadog_ai_gateway_routing_rule" "coding" {
  slug = "coding"

  tiers = [
    {
      ordering = "weighted_affinity"
      targets = [
        { provider = "prod-us", model = "gpt-5.2-codex", weight = 70 },
        { provider = "prod-us", model = "gpt-4o-mini", weight = 30 },
      ]
    },
    {
      targets = [{ routing_rule = "general" }]
    },
  ]
}
