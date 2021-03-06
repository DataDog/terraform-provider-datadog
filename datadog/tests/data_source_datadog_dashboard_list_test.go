package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDatadogDashboardListDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDashListDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDashboardListNameFilterConfig(uniq),
				Check:  checkDatasourceDashboardListAttrs(accProvider, uniq),
			},
		},
	})
}

func checkDatasourceDashboardListAttrs(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
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
