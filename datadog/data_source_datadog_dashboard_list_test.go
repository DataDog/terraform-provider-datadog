package datadog

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDatadogDashboardListDatasource(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	uniq := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
