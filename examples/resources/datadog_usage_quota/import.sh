# Import one quota owned by the organization authenticated by the provider.
# Obtain the opaque quota ID from the usage quotas API; do not decode or alter it.
terraform import datadog_usage_quota.example 'ai_credits/<quota_id>'
