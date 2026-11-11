resource "datadog_ai_gateway_provider" "openai" {
  name                   = "OpenAI production"
  slug                   = "openai-prod"
  known_provider         = "openai"
  integration_account_id = "<integration_account_id>"
  cost_multiplier        = 1.2
}

resource "datadog_ai_gateway_provider" "bedrock" {
  name                   = "Bedrock us-east-1"
  slug                   = "bedrock-us-east-1"
  known_provider         = "bedrock"
  integration_account_id = "<integration_account_id>"

  integration_config = {
    bedrock_region = "us-east-1"
  }
}
