package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardHeatMapConfig = `
resource "datadog_dashboard" "heatmap_dashboard" {
    title         = "Acceptance Test Heatmap Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "ordered"
    is_read_only  = "true"

    widget {
		heatmap_definition {
			title = "Avg of system.cpu.user over account:prod by app"
            title_align = "center"
			title_size = "16"
			yaxis {
				max = "100"
			}
			request {
				q = "avg:system.cpu.user{account:prod} by {app}"
				style {
					palette = "blue"
				}
			}
			
			time = {
				live_span = "1mo"
			}
			//event {
			//	q = "env:prod"
			//	tags_execution = "and"
			//}
			//legend_size = "2"
		}
    }
}
`

var datadogDashboardHeatMapAsserts = []string{
	"title = Acceptance Test Heatmap Widget Dashboard",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"is_read_only = true",
	"widget.0.heatmap_definition.0.title = Avg of system.cpu.user over account:prod by app",
	"widget.0.heatmap_definition.0.title_align = center",
	"widget.0.heatmap_definition.0.title_size = 16",
	"widget.0.heatmap_definition.0.request.0.q = avg:system.cpu.user{account:prod} by {app}",
	"widget.0.heatmap_definition.0.request.0.style.0.palette = blue",
	"widget.0.heatmap_definition.0.yaxis.0.include_zero = false",
	"widget.0.heatmap_definition.0.yaxis.0.label =",
	"widget.0.heatmap_definition.0.yaxis.0.max = 100",
	"widget.0.heatmap_definition.0.yaxis.0.scale =",
	"widget.0.heatmap_definition.0.yaxis.0.min =",
	"widget.0.heatmap_definition.0.time.live_span = 1mo",
}

func TestAccDatadogDashboardHeatMap(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardHeatMapConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.heatmap_dashboard", checkDashboardExists(accProvider), datadogDashboardHeatMapAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardHeatMap_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardHeatMapConfig,
			},
			{
				ResourceName:      "datadog_dashboard.heatmap_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
