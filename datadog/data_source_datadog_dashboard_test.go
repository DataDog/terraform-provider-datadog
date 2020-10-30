package datadog

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccDatadogDashboardDatasource(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	uniq := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:             testAccDatasourceDashboardNameFilterConfig(uniq),
				ExpectNonEmptyPlan: true,
				Check:              checkDatasourceDashboardAttrs(accProvider, uniq),
			},
		},
	})
}

func checkDatasourceDashboardAttrs(accProvider *schema.Provider, uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.datadog_dashboard.my_dash", "name", uniq),
		resource.TestCheckResourceAttrSet(
			"data.datadog_dashboard.my_dash", "id"),
		resource.TestCheckResourceAttrSet(
			"data.datadog_dashboard.my_dash", "url"),
		resource.TestCheckResourceAttrSet(
			"data.datadog_dashboard.my_dash", "title"),
	)
}

func testAccDashboardConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "foo" {
  title = "%s"
  layout_type = "ordered"
  widget {
	alert_graph_definition {
		alert_id = "895605"
		viz_type = "timeseries"
		title = "Widget Title"
		time = {
			live_span = "1h"
		}
	}
  }
}`, uniq)
}

func testAccDatasourceDashboardNameFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_dashboard" "my_dash" {
  depends_on = [
    datadog_dashboard.foo,
  ]
  name = "%s"
}`, testAccDashboardConfig(uniq), uniq)
}
