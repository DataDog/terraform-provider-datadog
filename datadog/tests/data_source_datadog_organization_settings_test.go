package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogOrganizationSettingsDatasource_basic(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogOrganizationSettingsDatasourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_organization_settings.test", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_organization_settings.test", "name"),
					resource.TestCheckResourceAttrSet("data.datadog_organization_settings.test", "public_id"),
				),
			},
		},
	})
}

func TestAccDatadogOrganizationSettingsDatasource_settings(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogOrganizationSettingsDatasourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify settings block exists
					resource.TestCheckResourceAttr("data.datadog_organization_settings.test", "settings.#", "1"),
					// Verify SAML block exists
					resource.TestCheckResourceAttr("data.datadog_organization_settings.test", "settings.0.saml.#", "1"),
					// Verify SAML autocreate users domains block exists
					resource.TestCheckResourceAttr("data.datadog_organization_settings.test", "settings.0.saml_autocreate_users_domains.#", "1"),
					// Verify SAML IdP initiated login block exists
					resource.TestCheckResourceAttr("data.datadog_organization_settings.test", "settings.0.saml_idp_initiated_login.#", "1"),
					// Verify SAML strict mode block exists
					resource.TestCheckResourceAttr("data.datadog_organization_settings.test", "settings.0.saml_strict_mode.#", "1"),
					// Verify boolean attributes are set
					resource.TestCheckResourceAttrSet("data.datadog_organization_settings.test", "settings.0.private_widget_share"),
					resource.TestCheckResourceAttrSet("data.datadog_organization_settings.test", "settings.0.saml_can_be_enabled"),
					resource.TestCheckResourceAttrSet("data.datadog_organization_settings.test", "settings.0.saml_idp_metadata_uploaded"),
				),
			},
		},
	})
}

func TestAccDatadogOrganizationSettingsDatasource_outputs(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogOrganizationSettingsDatasourceOutputConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source attributes are set
					resource.TestCheckResourceAttrSet("data.datadog_organization_settings.test", "name"),
					resource.TestCheckResourceAttrSet("data.datadog_organization_settings.test", "public_id"),
					// Verify settings can be accessed in outputs
					resource.TestCheckResourceAttrSet("data.datadog_organization_settings.test", "settings.0.saml.0.enabled"),
				),
			},
		},
	})
}

func testAccDatadogOrganizationSettingsDatasourceConfig() string {
	return `
data "datadog_organization_settings" "test" {}
`
}

func testAccDatadogOrganizationSettingsDatasourceOutputConfig() string {
	return `
data "datadog_organization_settings" "test" {}

output "org_name" {
  value = data.datadog_organization_settings.test.name
}

output "org_public_id" {
  value = data.datadog_organization_settings.test.public_id
}

output "saml_enabled" {
  value = data.datadog_organization_settings.test.settings[0].saml[0].enabled
}
`
}
