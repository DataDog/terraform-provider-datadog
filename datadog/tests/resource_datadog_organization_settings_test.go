package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func preludeAccDatadogOrganizationSettings(t *testing.T) {
	// Temporarily comment this whole function to update the cassettes recording
	if !isReplaying() {
		t.Skip("This test only supports replaying")
	}
	t.Parallel()
}

func TestAccDatadogOrganizationSettings_Update(t *testing.T) {
	preludeAccDatadogOrganizationSettings(t)
	ctx, accProviders := testAccProviders(context.Background(), t)
	organizationName := uniqueOrgName(ctx, "update")
	organizationName2 := uniqueOrgName(ctx, "updat2")
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

func TestAccDatadogOrganizationSettings_Import(t *testing.T) {
	preludeAccDatadogOrganizationSettings(t)
	resourceName := "datadog_organization_settings.foo"
	ctx, accProviders := testAccProviders(context.Background(), t)
	organizationName := uniqueOrgName(ctx, "import")

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

func TestAccDatadogOrganizationSettings_SecurityContacts(t *testing.T) {
	preludeAccDatadogOrganizationSettings(t)
	ctx, accProviders := testAccProviders(context.Background(), t)
	organizationName := uniqueOrgName(ctx, "secCon")
	cfg := func(extra string) string {
		return fmt.Sprintf(`
			resource datadog_organization_settings foo {
				name = "%s"
				%s
			}
		`, organizationName, extra)
	}

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      cfg(`security_contacts = [null]`),
				ExpectError: regexp.MustCompile("Error: Null value found in list"),
			},
			{
				Config:      cfg(`security_contacts = ["foo@example@org"]`),
				ExpectError: regexp.MustCompile("Error: not a email: foo@example@org"),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: cfg(``),
			},
			{
				Config: cfg(`security_contacts = ["foo@example.org"]`),
			},
			{
				Config: cfg(`security_contacts = ["foo@example.org", "bar@example.org"]`),
			},
			{
				Config:   cfg(`security_contacts = ["foo@example.org", "bar@example.org"]`),
				PlanOnly: true, // no change
			},
			{
				ResourceName:      "datadog_organization_settings.foo",
				Config:            cfg(``),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   cfg(`security_contacts = null`),
				PlanOnly: true, // not configured, no change
			},
			{
				Config:   cfg(``),
				PlanOnly: true, // not configured, no change
			},
		},
	})

	// Starting from scratch, unset the security contacts (test an edge case around zero-values plans)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: cfg(`security_contacts = []`),
			},
			{
				Config:   cfg(`security_contacts = []`),
				PlanOnly: true,
			},
		},
	})
}

func TestAccDatadogOrganizationSettings_IncorrectName(t *testing.T) {
	preludeAccDatadogOrganizationSettings(t)
	_, accProviders := testAccProviders(context.Background(), t)
	organizationName := "very-long-and-boring-name-longer-than-32-characters"

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

// uniqueOrgName returns an organization name for a particular test, whose id must be 6 char long at most.
// uniqueEntityName is not suitable because org names are limited to 32 characters.
func uniqueOrgName(ctx context.Context, id string) string {
	r := fmt.Sprintf("tf-plugin-test-%s-%d", id, clockFromContext(ctx).Now().Unix()%1e9)
	if len(r) > 32 {
		panic("uniqueOrgName is too long: " + r)
	}
	return r
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
