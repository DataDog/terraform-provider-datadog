# Source the permissions for scoped keys
data "datadog_permissions" "dd_perms" {}

# Create an unrestricted Service Account Application Key
# This key inherits all permissions of the service account that owns the key
resource "datadog_service_account_application_key" "unrestricted_key" {
  service_account_id = "00000000-0000-1234-0000-000000000000"
  name               = "Application key for managing dashboards"
  # scopes unset - inherits all service account permissions
}

# Create a scoped Service Account Application Key for monitor management
resource "datadog_service_account_application_key" "monitor_management_key" {
  service_account_id = "00000000-0000-1234-0000-000000000000"
  name               = "Monitor Management Service Account Key"
  scopes = [
    data.datadog_permissions.dd_perms.permissions.monitors_read,
    data.datadog_permissions.dd_perms.permissions.monitors_write
  ]
}

# Create new service_account_application_key resource with Actions API access (Preview)
resource "datadog_service_account_application_key" "actions_enabled" {
  service_account_id        = "00000000-0000-1234-0000-000000000000"
  name                      = "Application key for managing dashboards with Actions API"
  enable_actions_api_access = true
}
