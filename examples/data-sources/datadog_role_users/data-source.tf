# Get the API Key Manager role
data "datadog_role" "api_key_manager" {
  filter = "API Key Manager"
}

# List users assigned to the API Key Manager role
data "datadog_role_users" "api_key_managers" {
  role_id = data.datadog_role.api_key_manager.id
}