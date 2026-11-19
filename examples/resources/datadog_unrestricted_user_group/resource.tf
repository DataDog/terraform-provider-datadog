resource "datadog_role" "data_admins" {
  name = "Data Admins"
}

# Give members of the role unrestricted access to logs and APM data, and to all dashboards.
resource "datadog_unrestricted_user_group" "data_admins" {
  principal_id = "role:${datadog_role.data_admins.id}"
  products     = ["logs", "apm", "dashboard"]
}
