package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogDashboardDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDashboardNameFilterConfig(uniq),
				Check:  checkDatasourceDashboardAttrs(accProvider, uniq),
			},
		},
	})
}

func checkDatasourceDashboardAttrs(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.datadog_dashboard.dash_one", "name", uniq+" one"),
		resource.TestCheckResourceAttrSet(
			"data.datadog_dashboard.dash_one", "id"),
		resource.TestCheckResourceAttrSet(
			"data.datadog_dashboard.dash_one", "url"),
		resource.TestCheckResourceAttr(
			"data.datadog_dashboard.dash_one", "title", uniq+" one"),
	)
}

func testAccDashboardConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "dash_one" {
  title = "%s one"
  layout_type = "ordered"
  widget {
	alert_graph_definition {
		alert_id = "895605"
		viz_type = "timeseries"
		title = "Widget Title"
		live_span = "1h"
	}
  }
}
  resource "datadog_dashboard" "dash_two" {
	title = "%s two"
	layout_type = "ordered"
	widget {
	  alert_graph_definition {
		  alert_id = "895605"
		  viz_type = "timeseries"
		  title = "Widget Title"
		  live_span = "1h"
	  }
	}
}`, uniq, uniq)
}

func testAccDatasourceDashboardNameFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_dashboard" "dash_one" {
  depends_on = [
    datadog_dashboard.dash_one,
    datadog_dashboard.dash_two,
  ]
  name = "%s one"
}`, testAccDashboardConfig(uniq), uniq)
}
