package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// samlIdpMetadataTestCertificate is a long-lived self-signed certificate used
// only to build valid IdP metadata documents for this test.
const samlIdpMetadataTestCertificate = `MIIDPTCCAiWgAwIBAgIUfEQrXK0wPqsFvbULPeR1nrmhW1kwDQYJKoZIhvcNAQELBQAwLjEsMCoG
A1UEAwwjdGVycmFmb3JtLXByb3ZpZGVyLWRhdGFkb2ctdGVzdC1pZHAwHhcNMjYwODEwMDc0OTI1
WhcNNDYwODA1MDc0OTI1WjAuMSwwKgYDVQQDDCN0ZXJyYWZvcm0tcHJvdmlkZXItZGF0YWRvZy10
ZXN0LWlkcDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAJ6LiCfuesYVHXEJfAsW/qiV
YbeZffE6bg9r3zcSsipuUggQi5B//WpqZM3pQ+AXfMy8uT9mMBXBVqrWTyuHVXGABrN5f8bfmYUf
RZRXPNxSnkD0VKKCkngh1X+t2/FAayWFWsDAZSsTkCaXeHWGWPJUqr1fuOfUpCM6VYCwtu//smCb
UoFuucNhr5P1oeCNLTutpFrvnnXlZkVLjZ8Y+46Q/mR8SAO5c3fcqVoEWINvVv6+AjZuGs5e52vZ
SBWXX9uNxNkx/aQl1m/O7+3KzYXichz14cEOoA6a4GRTWFa8YV3bHQmjCtnNmmw1nILoRHe3v1cB
UFHzmNZDVq7GXRMCAwEAAaNTMFEwHQYDVR0OBBYEFDqp/IcqIHgOd8Y6VlrhQdSDgeZQMB8GA1Ud
IwQYMBaAFDqp/IcqIHgOd8Y6VlrhQdSDgeZQMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQEL
BQADggEBACdo68cb8Trymr8skrZZ2HTLIIsOGIEDh5IlrnYpsfwnry/W+ngquiIYG0xQYdjCrm08
9US2F/eYy4itd+SGc86LbcQOJDmqxX0oldSkR4UIFZFlKLvvLDRJNRhMtzFsTwfI08WpBVXxnRTz
oXFBib6uvsOTnrC3dsSsrx4ZyBiallCvXDRChhBI6HIElgVJsHV40Tu1Wnm1kesKl0nuSw3CSt1E
g8LXvsjmbx2BPFT9baM1JerhrkSGRFaANN/awmA+cKkU4/dpYuSVafkl2H2KfJnwb8GUZTR0ScMf
4+JHOY3PuNycVqOTo9HbQ7biasPmFaUP18l4NaCGG+dc4gc=`

// TestAccSamlIdpMetadataBasic is not parallel because the SAML configuration
// is an organization-level singleton.
func TestAccSamlIdpMetadataBasic(t *testing.T) {
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	entityId := fmt.Sprintf("https://idp.example.com/%s", uniq)
	ssoUrl := fmt.Sprintf("https://idp.example.com/%s/sso", uniq)
	ssoUrlUpdated := fmt.Sprintf("https://idp.example.com/%s/sso2", uniq)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSamlIdpMetadata(entityId, ssoUrl),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"datadog_saml_idp_metadata.foo", "id"),
					// entity_id refers to Datadog's service provider configuration, not
					// the uploaded metadata, so only assert that it is set.
					resource.TestCheckResourceAttrSet(
						"datadog_saml_idp_metadata.foo", "entity_id"),
					resource.TestCheckResourceAttrSet(
						"datadog_saml_idp_metadata.foo", "expires_at"),
				),
			},
			{
				Config: testAccCheckDatadogSamlIdpMetadata(entityId, ssoUrlUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"datadog_saml_idp_metadata.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_saml_idp_metadata.foo", "entity_id"),
				),
			},
		},
	})
}

func testAccCheckDatadogSamlIdpMetadata(entityId string, ssoUrl string) string {
	return fmt.Sprintf(`
resource "datadog_saml_idp_metadata" "foo" {
  idp_metadata = <<-EOF
<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" entityID="%s">
  <md:IDPSSODescriptor WantAuthnRequestsSigned="false" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:KeyDescriptor use="signing">
      <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
        <ds:X509Data>
          <ds:X509Certificate>%s</ds:X509Certificate>
        </ds:X509Data>
      </ds:KeyInfo>
    </md:KeyDescriptor>
    <md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:persistent</md:NameIDFormat>
    <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="%s"/>
    <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="%s"/>
  </md:IDPSSODescriptor>
</md:EntityDescriptor>
EOF
}
`, entityId, samlIdpMetadataTestCertificate, ssoUrl, ssoUrl)
}
