# Organization-wide policy: the inbound pattern
terraform import datadog_ai_gateway_policy.coding_agent "coding-agent"

# User-scoped policy: <user_id>:<inbound>
terraform import datadog_ai_gateway_policy.coding_agent "55555555-5555-5555-5555-555555555555:coding-agent"
