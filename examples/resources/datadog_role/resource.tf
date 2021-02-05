# Source the permissions
data "datadog_permissions" "bar" {}

# Create a new Datadog role
resource "datadog_role" "foo" {
  name = "foo"
  permission {
    id = data.datadog_permissions.bar.permissions.monitors_downtime
  }
  permission {
    id = data.datadog_permissions.bar.permissions.monitors_write
  }
}
