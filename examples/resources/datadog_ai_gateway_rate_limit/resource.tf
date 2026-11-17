data "datadog_team" "ml_platform" {
  filter_keyword = "ml-platform"
}

# Cap the ML platform team's OpenAI traffic.
resource "datadog_ai_gateway_rate_limit" "ml_platform_openai" {
  name                    = "ml-platform-openai"
  mode                    = "ENFORCE"
  requests_per_minute     = 600
  input_tokens_per_minute = 2000000
  team_id                 = data.datadog_team.ml_platform.id
  tags = {
    provider = "openai"
  }
}
