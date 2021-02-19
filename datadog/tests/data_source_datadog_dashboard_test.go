package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccDatadogDashboardDatasource(t *testing.T) {
	t.Parallel()
	ctx := testSpan(context.Background(), t)
	ctx, accProviders := testAccProviders(ctx, t, initRecorder(t))
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
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
		time = {
			live_span = "1h"
		}
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
		  time = {
			  live_span = "1h"
		  }
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
