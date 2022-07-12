# Source a role
data "datadog_role" "ro_role" {
  filter = "Datadog Read Only Role"
}

# Create a new Datadog user
resource "datadog_user" "foo" {
  email = "new@example.com"

  roles = [data.datadog_role.ro_role.id]
}

# Create a new Datadog service account
resource "datadog_user" "bar" {
  email = "new@example.com"

  roles = [data.datadog_role.ro_role.id]

  service_account = true
}
