package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogChildOrganization_Create(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying")
	}
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueEntity := uniqueEntityName(ctx, t)
	organizationName := fmt.Sprint(uniqueEntity[len(uniqueEntity)-32:])
	resourceName := "datadog_child_organization.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogChildOrganizationConfigRequired(organizationName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", organizationName),
					resource.TestCheckResourceAttrSet(resourceName, "public_id"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.private_widget_share"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_access_role"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_autocreate_users_domains.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_can_be_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_initiated_login.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_idp_metadata_uploaded"),
					resource.TestCheckResourceAttrSet(resourceName, "settings.0.saml_strict_mode.0.enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "api_key.0.key"),
					resource.TestCheckResourceAttrSet(resourceName, "api_key.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "application_key.0.hash"),
					resource.TestCheckResourceAttrSet(resourceName, "application_key.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "application_key.0.owner"),
				),
			},
		},
	})
}

func TestAccDatadogChildOrganization_IncorrectName(t *testing.T) {
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
				Config:      testAccCheckDatadogChildOrganizationConfigRequired(organizationName),
				ExpectError: regexp.MustCompile("expected length of name to be in the range \\(1 - 32\\)."),
			},
		},
	})
}

func testAccCheckDatadogChildOrganizationConfigRequired(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_child_organization" "foo" {
  name = "%s"
}`, uniq)
}
