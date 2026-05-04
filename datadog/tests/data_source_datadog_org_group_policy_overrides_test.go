package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogOrgGroupPolicyOverridesDataSource_Basic(t *testing.T) {
	// Not parallel: uses the shared test org's membership.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	dsAll := "data.datadog_org_group_policy_overrides.all"
	dsFiltered := "data.datadog_org_group_policy_overrides.filtered"

	orgUUID := getTestOrgUUID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances)
	originalGroupID := getOrgCurrentGroupID(providers.frameworkProvider.Auth, t, providers.frameworkProvider.DatadogApiInstances, orgUUID)

	t.Cleanup(restoreOrgMembership(t, providers.frameworkProvider, orgUUID, originalGroupID))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             composeOrgGroupStackDestroyChecks(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupPolicyOverridesDataSourceConfig(orgGroupName, orgUUID),
				Check: resource.ComposeTestCheckFunc(
					// Unfiltered list contains our explicit override.
					resource.TestCheckResourceAttr(dsAll, "overrides.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(dsAll, "overrides.*", map[string]string{
						"org_uuid": orgUUID,
						"org_site": overrideTestOrgSite,
					}),
					// Client-side org_uuid filter keeps the matching row.
					resource.TestCheckResourceAttr(dsFiltered, "overrides.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(dsFiltered, "overrides.*", map[string]string{
						"org_uuid": orgUUID,
					}),
				),
			},
			// Restore original membership so the test org_group can be destroyed.
			{
				Config: testAccCheckDatadogOrgGroupPolicyOverridesDataSourceRestore(orgGroupName, orgUUID, originalGroupID),
			},
		},
	})
}

func testAccCheckDatadogOrgGroupPolicyOverridesDataSourceConfig(orgGroupName, orgUUID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = datadog_org_group.foo.id
  org_uuid     = "%s"
}

resource "datadog_org_group_policy" "foo" {
  org_group_id     = datadog_org_group.foo.id
  policy_name      = "is_widget_copy_paste_enabled"
  content          = jsonencode({"org_config": false})
  enforcement_tier = "DEFAULT"
}

resource "datadog_org_group_policy_override" "foo" {
  org_group_id = datadog_org_group.foo.id
  policy_id    = datadog_org_group_policy.foo.id
  org_uuid     = "%s"
  org_site     = "%s"
  depends_on   = [datadog_org_group_membership.foo]
}

data "datadog_org_group_policy_overrides" "all" {
  org_group_id = datadog_org_group.foo.id
  depends_on   = [datadog_org_group_policy_override.foo]
}

data "datadog_org_group_policy_overrides" "filtered" {
  org_group_id = datadog_org_group.foo.id
  policy_id    = datadog_org_group_policy.foo.id
  org_uuid     = "%s"
  depends_on   = [datadog_org_group_policy_override.foo]
}`, orgGroupName, orgUUID, orgUUID, overrideTestOrgSite, orgUUID)
}

func testAccCheckDatadogOrgGroupPolicyOverridesDataSourceRestore(orgGroupName, orgUUID, originalGroupID string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

resource "datadog_org_group_membership" "foo" {
  org_group_id = "%s"
  org_uuid     = "%s"
}`, orgGroupName, originalGroupID, orgUUID)
}
