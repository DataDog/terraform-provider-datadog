package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogOrgGroupsDataSource_Basic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	orgGroupName := uniqueEntityName(ctx, t)
	dsName := "data.datadog_org_groups.all"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgGroupsDataSourceConfig(orgGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dsName, "groups.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					// The list must contain the org group we just created.
					resource.TestCheckTypeSetElemNestedAttrs(dsName, "groups.*", map[string]string{
						"name": orgGroupName,
					}),
				),
			},
		},
	})
}

func testAccCheckDatadogOrgGroupsDataSourceConfig(orgGroupName string) string {
	return fmt.Sprintf(`
resource "datadog_org_group" "foo" {
  name = "%s"
}

data "datadog_org_groups" "all" {
  depends_on = [datadog_org_group.foo]
}`, orgGroupName)
}
