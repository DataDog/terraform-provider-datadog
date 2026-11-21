# Web Integration Account examples for Twilio, Snowflake, and Databricks
#
# The settings_json and secrets_json structure varies by integration.
# Use GET /api/v2/web-integrations/{integration_name}/accounts/schema
# to retrieve the schema for your integration.
#
# Schema reference: dd-source/domains/web-integrations/shared/libs/go/schemas/schemas/

# --- Twilio ---
resource "datadog_web_integration_account" "twilio" {
  integration_name = "twilio"
  name             = "My Twilio Production Account"

  settings_json = jsonencode({
    api_key       = "SKxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    account_sid   = "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    events        = true
    messages      = true
    alerts        = true
    call_summaries = true
    ccm_enabled   = true
    censor_logs   = true
  })

  secrets_json = jsonencode({
    api_key_token = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  })
}

# --- Snowflake (snowflake-web) ---
resource "datadog_web_integration_account" "snowflake" {
  integration_name = "snowflake-web"
  name             = "My Snowflake Account"

  settings_json = jsonencode({
    username                     = "DATADOG_MONITOR_USER"
    snowflake_account_identifier = "myorg-account123.us-east-1.aws"
    private_key_name             = "datadog_rsa_key"
    query_tags                   = "env,team,cost_center"
    do_table_crawler_cron        = "0 * * * *"

    # Logs and traces
    query_history_logs_enabled                = true
    task_history_logs_enabled                 = true
    task_history_traces_enabled                = true
    join_query_history_with_access_history_enabled = true
    event_table_logs_enabled                   = true
    event_table_events_enabled                 = true
    login_history_logs_enabled                 = true
    sessions_logs_enabled                      = true

    # Metrics
    account_usage_metrics_enabled              = true
    organization_usage_metrics_enabled         = true
    event_table_metrics_enabled                = false
    account_usage_metrics_aggregate_last_24h    = false
    organization_usage_metrics_aggregate_last_24h = false

    # Data Observability and CCM
    datasets_enabled             = true
    ccm_enabled                  = true
    ccm_only_monitor_account     = true
    ccm_has_orgadmin             = true
    ccm_credit_dollar_override   = 2.6
    ccm_terrabyte_month_override = 23.0
    ccm_query_tags               = ["env", "team"]

    # Other features
    grants_to_users_enabled      = true
    data_transfer_history_enabled = true
    stages_enabled              = true
    snowpark_traces              = false
    product_analytics_enabled   = false
    data_observability_enabled  = true

    # Collection intervals (minutes)
    security_logs_interval_min        = 5
    query_history_logs_interval_min   = 15
    task_history_logs_interval_min    = 15
    task_history_traces_interval_min  = 15
    event_table_logs_interval_min     = 15
  })

  secrets_json = jsonencode({
    private_key = "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBg...\n-----END PRIVATE KEY-----"
  })
}

# --- Databricks (OAuth, preferred) ---
resource "datadog_web_integration_account" "databricks" {
  integration_name = "databricks"
  name             = "my-databricks-workspace"

  settings_json = jsonencode({
    workspace_url = "https://my-workspace.cloud.databricks.com"

    # OAuth authentication (preferred over token)
    client_id             = "my-client-id"
    databricks_account_id = "my-databricks-account-id"

    # Datadog API key pair (for lineage/DO event ingestion)
    dd_api_key_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

    # System tables (required for Cloud Cost Management)
    system_tables_sql_warehouse_id = "my-warehouse-id"
    model_serving_endpoint_name   = "my-model-endpoint"

    # Product toggles
    djm_enabled                    = true
    djm_global_init_script_enabled = false
    ccm_enabled                    = true
    do_enabled                     = true
    do_crawlers_cron               = "0 * * * *"
    model_serving_metrics_enabled  = true
    script_logs_enabled            = true
    script_gpum_enabled            = false
    table_lineage_enabled          = true
    serverless_jobs_enabled        = true

    # Private Action Runner (for executing actions in Databricks)
    private_action_runner_configuration = {
      connection_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      user_uuid     = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
      secret_path   = "path/to/databricks/credentials"
    }
  })

  secrets_json = jsonencode({
    client_secret     = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    dd_api_key_secret = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  })
}

# --- Databricks (legacy: Personal Access Token) ---
resource "datadog_web_integration_account" "databricks_legacy" {
  integration_name = "databricks"
  name             = "my-legacy-workspace"

  settings_json = jsonencode({
    workspace_url                  = "https://legacy-workspace.cloud.databricks.com"
    dd_api_key_id                  = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    system_tables_sql_warehouse_id = "my-warehouse-id"
    djm_enabled                    = true
    djm_global_init_script_enabled = false
    ccm_enabled                    = true
    do_enabled                     = false
    do_crawlers_cron               = "0 * * * *"
    model_serving_metrics_enabled  = false
    script_logs_enabled            = false
    script_gpum_enabled            = false
    table_lineage_enabled          = false
    serverless_jobs_enabled        = true
  })

  secrets_json = jsonencode({
    token           = "dapixxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    dd_api_key_secret = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  })
}
