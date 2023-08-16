# Create a new Datadog - Microsoft Azure integration
resource "datadog_integration_azure" "sandbox" {
  tenant_name              = "<azure_tenant_name>"
  client_id                = "<azure_client_id>"
  client_secret            = "<azure_client_secret_key>"
  host_filters             = "examplefilter:true,example:true"
  app_service_plan_filters = "examplefilter:true,example:another"
  automute                 = true
  cspm_enabled             = true
  custom_metrics_enabled   = false
}
