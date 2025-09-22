# Create a new Datadog Application Key
resource "datadog_application_key" "foo" {
  name = "foo-application"
}

# Create a new Datadog Application Key with Actions API access (Preview)
resource "datadog_application_key" "actions_enabled" {
  name                      = "foo-application-with-actions"
  enable_actions_api_access = true
}
