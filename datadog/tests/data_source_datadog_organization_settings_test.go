package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogOrganizationSettingsDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)
	resourceName := "data.datadog_organization_settings.organization"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationSettingsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "public_id"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.private_widget_share"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_access_role"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_can_be_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_initiated_login.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_metadata_uploaded"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_strict_mode.0.enabled"),
				),
			},
		},
	})
}

const testAccOrganizationSettingsConfig = `
data "datadog_organization_settings" "organization" {
}`
