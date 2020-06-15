package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardAlertGraphConfig = `
resource "datadog_dashboard" "alert_graph_dashboard" {
    title         = "Acceptance Test Alert Graph Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "ordered"
    is_read_only  = true
    widget {
		alert_graph_definition {
			alert_id = "895605"
			viz_type = "timeseries"
		}
    }
    widget {
		alert_graph_definition {
			alert_id = "895606"
			viz_type = "toplist"
			title = "Widget Title"
            title_align = "right"
			title_size = "16"
			time = {
				live_span = "1h"
			}
		}
    }
}
`

func TestAccDatadogDashboardAlertGraph(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardAlertGraphConfig,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					// Dashboard metadata
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "title", "Acceptance Test Alert Graph Widget Dashboard"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "description", "Created using the Datadog provider in Terraform"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "layout_type", "ordered"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "is_read_only", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "widget.#", "2"),
					// Alert Graph widget
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "widget.0.alert_graph_definition.0.alert_id", "895605"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "widget.0.alert_graph_definition.0.viz_type", "timeseries"),

					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "widget.1.alert_graph_definition.0.alert_id", "895606"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "widget.1.alert_graph_definition.0.viz_type", "toplist"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "widget.1.alert_graph_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "widget.1.alert_graph_definition.0.title_align", "right"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "widget.1.alert_graph_definition.0.title_size", "16"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_graph_dashboard", "widget.1.alert_graph_definition.0.time.live_span", "1h"),
				),
			},
		},
	})
}
