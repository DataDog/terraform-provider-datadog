# Upload the SAML IdP metadata for the organization from a local file
resource "datadog_saml_idp_metadata" "example" {
  idp_metadata = file("${path.module}/idp_metadata.xml")
}
