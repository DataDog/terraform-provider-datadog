data "datadog_permissions" "bar" {}

resource "datadog_role" "monitor_writer_role" {
  name = "Monitor Writer Role"
}

# Create new user_role resource
resource "datadog_role_permission" "new_role_with_monitor_write_permission" {
  role_id       = datadog_role.monitor_writer_role.id
  permission_id = data.datadog_permissions.bar.permissions.monitors_write
}