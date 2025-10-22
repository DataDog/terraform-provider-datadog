# Source the permissions
data "datadog_permissions" "dd_perms" {}

# Create an API Key Manager role
resource "datadog_role" "api_key_manager" {
  name = "API Key Manager"
  permission {
    id = data.datadog_permissions.dd_perms.permissions.api_keys_read
  }
  permission {
    id = data.datadog_permissions.dd_perms.permissions.api_keys_write
  }
}

resource "datadog_user" "new_user" {
  email = "new@example.com"
}

# Assign the API Key Manager role to the user
resource "datadog_user_role" "new_user_with_api_key_manager_role" {
  role_id = datadog_role.api_key_manager.id
  user_id = datadog_user.new_user.id
}