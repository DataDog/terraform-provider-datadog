package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const unrestrictedUserGroupResourceName = "datadog_unrestricted_user_group.foo"

func TestAccUnrestrictedUserGroup(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	roleName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogUnrestrictedUserGroupDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccUnrestrictedUserGroupConfig(roleName, `["logs", "monitor"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUnrestrictedUserGroupExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrPair(unrestrictedUserGroupResourceName, "id", unrestrictedUserGroupResourceName, "principal_id"),
					resource.TestCheckResourceAttr(unrestrictedUserGroupResourceName, "products.#", "2"),
					resource.TestCheckTypeSetElemAttr(unrestrictedUserGroupResourceName, "products.*", "logs"),
					resource.TestCheckTypeSetElemAttr(unrestrictedUserGroupResourceName, "products.*", "monitor"),
				),
			},
			{
				// Products omitted from the request are removed from the group.
				Config: testAccUnrestrictedUserGroupConfig(roleName, `["apm", "dashboard", "logs"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogUnrestrictedUserGroupExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(unrestrictedUserGroupResourceName, "products.#", "3"),
					resource.TestCheckTypeSetElemAttr(unrestrictedUserGroupResourceName, "products.*", "apm"),
					resource.TestCheckTypeSetElemAttr(unrestrictedUserGroupResourceName, "products.*", "dashboard"),
					resource.TestCheckTypeSetElemAttr(unrestrictedUserGroupResourceName, "products.*", "logs"),
				),
			},
			{
				ResourceName:      unrestrictedUserGroupResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUnrestrictedUserGroupConfig(roleName, products string) string {
	return fmt.Sprintf(`
resource "datadog_role" "foo" {
  name = "%s"
}

resource "datadog_unrestricted_user_group" "foo" {
  principal_id = "role:${datadog_role.foo.id}"
  products     = %s
}`, roleName, products)
}

func testAccCheckDatadogUnrestrictedUserGroupExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return unrestrictedUserGroupExistsHelper(accProvider, s, true)
	}
}

func testAccCheckDatadogUnrestrictedUserGroupDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return unrestrictedUserGroupExistsHelper(accProvider, s, false)
	}
}

func unrestrictedUserGroupExistsHelper(accProvider *fwprovider.FrameworkProvider, s *terraform.State, shouldExist bool) error {
	apiInstances := accProvider.DatadogApiInstances
	auth := accProvider.Auth

	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_unrestricted_user_group" {
			continue
		}
		resp, httpResp, err := apiInstances.GetRestrictionPoliciesApiV2().ListUnrestrictedUserGroups(auth)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error listing unrestricted user groups")
		}
		found := false
		for _, group := range resp.GetData() {
			if group.GetId() == r.Primary.ID {
				found = true
				break
			}
		}
		if found != shouldExist {
			if shouldExist {
				return fmt.Errorf("unrestricted user group %s not found", r.Primary.ID)
			}
			return fmt.Errorf("unrestricted user group %s still exists", r.Primary.ID)
		}
	}
	return nil
}
