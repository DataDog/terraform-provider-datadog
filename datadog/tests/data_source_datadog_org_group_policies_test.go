package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogOrgGroupPoliciesDataSource_Basic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	dsAll := "data.datadog_org_group_policies.all"
	dsFiltered := "data.datadog_org_group_policies.filtered"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupPoliciesDataSourceConfig(orgGroupName),
				Check: resource.ComposeTestCheckFunc(
					// Two policies created, two in the unfiltered list.
					resource.TestCheckResourceAttr(dsAll, "policies.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(dsAll, "policies.*", map[string]string{
						"policy_name": "is_widget_copy_paste_enabled",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dsAll, "policies.*", map[string]string{
						"policy_name": "is_dashboard_reports_enabled",
					}),
					// Name filter returns only the matching policy.
					resource.TestCheckResourceAttr(dsFiltered, "policies.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(dsFiltered, "policies.*", map[string]string{
						"policy_name":      "is_widget_copy_paste_enabled",
						"enforcement_tier": "DEFAULT",
						"policy_type":      "org_config",
					}),
				),
			},
		},
	})
}

func testAccCheckDatadogOrgGroupPoliciesDataSourceConfig(orgGroupName string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

resource "datadog_org_group_policy" "widget" {
  org_group_id     = datadog_org_group.foo.id
  policy_name      = "is_widget_copy_paste_enabled"
  content          = jsonencode({"org_config": false})
  enforcement_tier = "DEFAULT"
}

resource "datadog_org_group_policy" "dashboard" {
  org_group_id     = datadog_org_group.foo.id
  policy_name      = "is_dashboard_reports_enabled"
  content          = jsonencode({"org_config": false})
  enforcement_tier = "DEFAULT"
}

data "datadog_org_group_policies" "all" {
  org_group_id = datadog_org_group.foo.id
  depends_on = [
    datadog_org_group_policy.widget,
    datadog_org_group_policy.dashboard,
  ]
}

data "datadog_org_group_policies" "filtered" {
  org_group_id = datadog_org_group.foo.id
  policy_name  = "is_widget_copy_paste_enabled"
  depends_on = [
    datadog_org_group_policy.widget,
    datadog_org_group_policy.dashboard,
  ]
}`, orgGroupName)
}
