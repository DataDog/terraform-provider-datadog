# Manage Datadog Organization
resource "datadog_organization_settings" "organization" {
  name = "foo-organization"
  saml_configurations {
    idp_metadata= file("/path/to/metadata.xml")
  }
}
