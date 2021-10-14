package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogOrganizationSettings_Update(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying")
	}
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueEntity := uniqueEntityName(ctx, t)
	organizationName := fmt.Sprint(uniqueEntity[len(uniqueEntity)-30:])
	organizationName2 := organizationName + "-2"
	resourceName := "datadog_organization_settings.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrganizationSettingsConfig_EmptySettings(organizationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", organizationName),
					resource.TestCheckResourceAttrSet(resourceName, "public_id"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.private_widget_share"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.private_widget_share", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_access_role"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_access_role", "st"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.domains.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_can_be_enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_can_be_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_initiated_login.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_idp_initiated_login.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_metadata_uploaded"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_idp_metadata_uploaded", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_strict_mode.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_strict_mode.0.enabled", "false"),
				),
			},
			{
				Config: testAccCheckDatadogOrganizationSettingsConfig_Required(organizationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", organizationName),
					resource.TestCheckResourceAttrSet(resourceName, "public_id"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.private_widget_share"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.private_widget_share", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_access_role"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_access_role", "st"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.domains.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_can_be_enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_can_be_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_initiated_login.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_idp_initiated_login.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_metadata_uploaded"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_idp_metadata_uploaded", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_strict_mode.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_strict_mode.0.enabled", "false"),
				),
			},
			{
				Config: testAccCheckDatadogOrganizationSettingsConfig_SomeParameters(organizationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", organizationName),
					resource.TestCheckResourceAttrSet(resourceName, "public_id"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.private_widget_share"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.private_widget_share", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_access_role"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_access_role", "st"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.domains.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.domains.0", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.domains.1", "www.example.com"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_can_be_enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_can_be_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_initiated_login.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_idp_initiated_login.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_metadata_uploaded"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_idp_metadata_uploaded", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_strict_mode.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_strict_mode.0.enabled", "false"),
				),
			},
			{
				Config: testAccCheckDatadogOrganizationSettingsConfig_Required(organizationName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", organizationName2),
					resource.TestCheckResourceAttrSet(resourceName, "public_id"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.private_widget_share"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.private_widget_share", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_access_role"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_access_role", "st"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.domains.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.domains.0", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_autocreate_users_domains.0.domains.1", "www.example.com"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_can_be_enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_can_be_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_initiated_login.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_idp_initiated_login.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_metadata_uploaded"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_idp_metadata_uploaded", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_strict_mode.0.enabled"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.saml_strict_mode.0.enabled", "false"),
				),
			},
		},
	})
}

func TestDatadogOrganizationSettings_import(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying")
	}
	t.Parallel()
	resourceName := "datadog_organization_settings.foo"
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueEntity := uniqueEntityName(ctx, t)
	organizationName := fmt.Sprint(uniqueEntity[len(uniqueEntity)-30:])

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrganizationSettingsConfig_Required(organizationName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogOrganizationSettings_IncorrectName(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying")
	}
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	organizationName := uniqueEntityName(ctx, t) + "qwertyuiopasdfghjklzxcvbnm0123456" // 32 characters

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogOrganizationSettingsConfig_Required(organizationName),
				ExpectError: regexp.MustCompile("expected length of name to be in the range \\(1 - 32\\)."),
			},
		},
	})
}

func testAccCheckDatadogOrganizationSettingsConfig_Required(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_organization_settings" "foo" {
  name = "%s"
}`, uniq)
}

func testAccCheckDatadogOrganizationSettingsConfig_EmptySettings(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_organization_settings" "foo" {
  name = "%s"

  settings {
	saml {
	}
	saml_autocreate_users_domains {
	}
	saml_idp_initiated_login {
	}
	saml_strict_mode {
	}
  }
}`, uniq)
}

func testAccCheckDatadogOrganizationSettingsConfig_SomeParameters(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_organization_settings" "foo" {
  name = "%s"

  settings {
    private_widget_share = true
	saml {
		enabled = false
	}
	saml_autocreate_users_domains {
      enabled = true
      domains = ["example.com", "www.example.com"]
	}
	saml_idp_initiated_login {
		enabled = false
	}
	saml_strict_mode {
		enabled = false
	}
  }
}`, uniq)
}
