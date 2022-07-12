# Create a new Datadog Application Key
resource "datadog_application_key" "foo" {
  name = "foo-application"
}

# Create a new Datadog Service Account Application Key
# Source a role
data "datadog_role" "ro_role" {
  filter = "Datadog Read Only Role"
}

# Create a new Datadog service account
resource "datadog_user" "foo" {
  email = "new@example.com"

  roles = [data.datadog_role.ro_role.id]

  service_account = true
}

resource "datadog_application_key" "bar" {
  name = "bar-application"

  service_account = datadog_user.foo.id
}
