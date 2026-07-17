# Get the API Key Manager role
data "datadog_role" "api_key_manager" {
  filter = "API Key Manager"
}

# List permissions assigned to the API Key Manager role
data "datadog_role_permissions" "api_key_manager_permissions" {
  role_id = data.datadog_role.api_key_manager.id
}
