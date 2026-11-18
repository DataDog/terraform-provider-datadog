# Route coding-agent traffic to Codex on the production provider, falling back to any provider serving Claude Sonnet.
resource "datadog_ai_gateway_policy" "coding_agent" {
  inbound = "coding-agent"

  tiers = [
    {
      targets = [{ provider = "prod-us", model = "gpt-5.2-codex" }]
    },
    {
      targets = [{ model = "claude-sonnet-5*" }]
    },
  ]
}
