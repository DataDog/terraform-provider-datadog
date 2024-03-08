resource "datadog_role" "foo" {
  name = "foo"
  permission {
    id = data.datadog_permissions.bar.permissions.monitors_downtime
  }
  permission {
    id = data.datadog_permissions.bar.permissions.monitors_write
  }
}

resource "datadog_user" "foo" {
  email = "new@example.com"
}

# Create new team_membership resource
resource "datadog_user_role" "foo" {
  team_id = datadog_role.foo.id
  user_id = datadog_user.foo.id
}