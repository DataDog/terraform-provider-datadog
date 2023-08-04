package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogDashboardListDatasource(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDashListDestroyWithFw(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDashboardListNameFilterConfig(uniq),
				Check:  checkDatasourceDashboardListAttrs(uniq),
			},
		},
	})
}

func checkDatasourceDashboardListAttrs(uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.datadog_dashboard_list.my_list", "name", uniq),
		resource.TestCheckResourceAttrSet(
			"data.datadog_dashboard_list.my_list", "id"),
	)
}

func testAccDashboardListConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_list" "foo" {
  name = "%s"
}`, uniq)
}

func testAccDatasourceDashboardListNameFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_dashboard_list" "my_list" {
  depends_on = [
    datadog_dashboard_list.foo,
  ]
  name = "%s"
}`, testAccDashboardListConfig(uniq), uniq)
}
