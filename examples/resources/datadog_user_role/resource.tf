resource "datadog_role" "monitor_writer_role" {
  name = "Monitor Writer Role"
  permission {
    id = data.datadog_permissions.bar.permissions.monitors_write
  }
}

resource "datadog_user" "new_user" {
  email = "new@example.com"
}

# Create new user_role resource
resource "datadog_user_role" "new_user_with_monitor_writer_role" {
  role_id = datadog_role.monitor_writer_role.id
  user_id = datadog_user.new_user.id
}