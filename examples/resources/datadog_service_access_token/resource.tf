# Source the permissions for scoped tokens
data "datadog_permissions" "dd_perms" {}

# Create a long-lived Service Access Token (SAT) for monitor management
resource "datadog_service_access_token" "monitor_management_token" {
  service_account_id = "00000000-0000-1234-0000-000000000000"
  name               = "Monitor Management Service Access Token"
  scopes = [
    data.datadog_permissions.dd_perms.permissions.monitors_read,
    data.datadog_permissions.dd_perms.permissions.monitors_write
  ]
}

# Create a Service Access Token (SAT) with a fixed expiration date.
# The Datadog API caps expirations at 365 days from creation. expires_at is
# immutable after creation; to rotate it, destroy and re-create the resource.
resource "datadog_service_access_token" "expiring_token" {
  service_account_id = "00000000-0000-1234-0000-000000000000"
  name               = "Short-lived CI Service Access Token"
  scopes = [
    data.datadog_permissions.dd_perms.permissions.dashboards_read
  ]
  expires_at = "2027-01-01T00:00:00Z"
}
