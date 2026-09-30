# Import using the composite ID format: integration_name:account_id
# The account_id is the UUID returned when creating the resource (e.g., from terraform state or API)

# Twilio
terraform import datadog_web_integration_account.twilio "twilio:abc123def456"

# Snowflake
terraform import datadog_web_integration_account.snowflake "snowflake-web:abc123def456"

# Databricks (OAuth)
terraform import datadog_web_integration_account.databricks "databricks:abc123def456"

# Databricks (legacy PAT)
terraform import datadog_web_integration_account.databricks_legacy "databricks:def456abc123"
