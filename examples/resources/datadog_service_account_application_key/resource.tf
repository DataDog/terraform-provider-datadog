# Create new service_account_application_key resource
resource "datadog_service_account_application_key" "foo" {
  service_account_id = "00000000-0000-1234-0000-000000000000"
  name               = "Application key for managing dashboards"
}

# Create new service_account_application_key resource with Actions API access (Preview)
resource "datadog_service_account_application_key" "actions_enabled" {
  service_account_id        = "00000000-0000-1234-0000-000000000000"
  name                      = "Application key for managing dashboards with Actions API"
  enable_actions_api_access = true
}