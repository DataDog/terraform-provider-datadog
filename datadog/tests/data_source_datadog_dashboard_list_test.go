package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccDatadogDashboardListDatasource(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(ctx, t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDashListDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:             testAccDatasourceDashboardListNameFilterConfig(uniq),
				ExpectNonEmptyPlan: true,
				Check:              checkDatasourceDashboardListAttrs(accProvider, uniq),
			},
		},
	})
}

func checkDatasourceDashboardListAttrs(accProvider *schema.Provider, uniq string) resource.TestCheckFunc {
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
