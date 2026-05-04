package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogOrgGroup_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	orgGroupNameUpdated := orgGroupName + "-updated"
	resourceName := "datadog_org_group.foo"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOrgGroupDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupConfig(orgGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", orgGroupName),
					resource.TestCheckResourceAttrSet(resourceName, "owner_org_site"),
					resource.TestCheckResourceAttrSet(resourceName, "owner_org_uuid"),
				),
			},
			{
				Config: testAccCheckDatadogOrgGroupConfig(orgGroupNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgGroupExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", orgGroupNameUpdated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogOrgGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}`, name)
}

func testAccCheckDatadogOrgGroupExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		id, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("org group ID is not a valid UUID: %w", err)
		}

		_, _, err = apiInstances.GetOrgGroupsApiV2().GetOrgGroup(auth, id)
		if err != nil {
			return fmt.Errorf("received an error retrieving org group: %w", err)
		}
		return nil
	}
}

func testAccCheckDatadogOrgGroupDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_org_group" {
				continue
			}

			id, err := uuid.Parse(r.Primary.ID)
			if err != nil {
				return fmt.Errorf("org group ID is not a valid UUID: %w", err)
			}

			_, httpResp, err := apiInstances.GetOrgGroupsApiV2().GetOrgGroup(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving org group: %w", err)
			}

			return fmt.Errorf("org group still exists")
		}

		return nil
	}
}
