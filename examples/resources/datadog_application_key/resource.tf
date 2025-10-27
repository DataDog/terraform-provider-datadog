# See the permissions available for scoped keys at https://docs.datadoghq.com/account_management/rbac/permissions/#permissions-list

# Create an unrestricted Application Key
# This key inherits all permissions of the user that owns the key
resource "datadog_application_key" "unrestricted_key" {
  name = "Unrestricted Application Key"
  # scopes unset - inherits all user permissions
}

# Create a scoped Application Key for monitor management
resource "datadog_application_key" "monitor_management_key" {
  name = "Monitor Management Key"
  scopes = [
    "monitors_read",
    "monitors_write"
  ]
}
