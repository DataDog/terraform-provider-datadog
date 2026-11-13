# Block the coding assistant's OpenAI spend over $500 a month.
resource "datadog_ai_gateway_budget" "coding_assistant" {
  name                = "coding-assistant-openai"
  capacity_microcents = 50000000000
  window_policy       = "MONTHLY"
  enforcement         = "BLOCK"
  warning_threshold   = 0.8
  tags = {
    ml_app   = "coding-assistant"
    provider = "openai"
  }
}

# Give each user their own $20 daily budget.
resource "datadog_ai_gateway_budget" "per_user" {
  name                = "per-user-daily"
  capacity_microcents = 2000000000
  window_policy       = "DAILY"
  enforcement         = "NOTIFY"
  selector = [
    { dim = "user_id", per_value = true },
  ]
}
