# Source the role
data "datadog_role" "ro_role" {
  filter = "Datadog Read Only Role"
}

# Create a new AuthN mapping
resource "datadog_authn_mapping" "dev_ro_role_mapping" {
  key   = "Member-of"
  value = "Development"
  role  = data.datadog_role.ro_role.id
}
