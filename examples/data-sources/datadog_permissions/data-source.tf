data "datadog_permissions" "dd_perms" {}

# Example of using specific permissions to create an API Key Manager role
resource "datadog_role" "api_key_manager" {
  name = "API Key Manager"
  permission {
    id = data.datadog_permissions.dd_perms.permissions.api_keys_read
  }
  permission {
    id = data.datadog_permissions.dd_perms.permissions.api_keys_write
  }
}

