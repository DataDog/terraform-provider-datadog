# Source a role
data "datadog_role" "ro_role" {
  filter = "Datadog Read Only Role"
}
# Create a new Datadog service account
resource "datadog_service_account" "bar" {
  email = "new@example.com"
  name = "Service Account Bar"
  roles = [data.datadog_role.ro_role.id]
}
