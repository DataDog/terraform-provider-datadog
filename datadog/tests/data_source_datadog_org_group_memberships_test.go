package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogOrgGroupMembershipsDataSource_Basic(t *testing.T) {
	// Not parallel: uses the shared test org's membership.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	dsName := "data.datadog_org_group_memberships.foo"

	orgUUID := getTestOrgUUID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances)
	originalGroupID := getOrgCurrentGroupID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances, orgUUID)

	t.Cleanup(restoreOrgMembership(t, providers.frameworkProvider, orgUUID, originalGroupID))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Move the test org into our new org_group and look it up by org_group_id.
				Config: testAccCheckDatadogOrgGroupMembershipsDataSourceConfig(orgGroupName, orgUUID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "memberships.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(dsName, "memberships.*", map[string]string{
						"org_uuid": orgUUID,
					}),
				),
			},
			// Restore the original membership so the test org_group can be destroyed.
			{
				Config: testAccCheckDatadogOrgGroupMembershipsDataSourceRestore(orgGroupName, orgUUID, originalGroupID),
			},
		},
	})
}

func testAccCheckDatadogOrgGroupMembershipsDataSourceConfig(orgGroupName, orgUUID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = datadog_org_group.foo.id
  org_uuid     = "%s"
}

data "datadog_org_group_memberships" "foo" {
  org_group_id = datadog_org_group.foo.id
  depends_on   = [datadog_org_group_membership.foo]
}`, orgGroupName, orgUUID)
}

func testAccCheckDatadogOrgGroupMembershipsDataSourceRestore(orgGroupName, orgUUID, originalGroupID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = "%s"
  org_uuid     = "%s"
}`, orgGroupName, originalGroupID, orgUUID)
}
