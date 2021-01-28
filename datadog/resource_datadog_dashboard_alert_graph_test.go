package datadog

import (
	"testing"
)

const datadogDashboardAlertGraphConfig = `
resource "datadog_dashboard" "alert_graph_dashboard" {
	title         = "{{uniq}}"
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
			live_span = "1h"
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

var datadogDashboardAlertGraphAsserts = []string{
	"title = {{uniq}}",
	"widget.0.alert_graph_definition.0.alert_id = 895605",
	"widget.1.alert_graph_definition.0.time.% = 0",
	"widget.1.alert_graph_definition.0.title = Widget Title",
	"is_read_only = true",
	"widget.1.alert_graph_definition.0.title_size = 16",
	"widget.1.alert_graph_definition.0.viz_type = toplist",
	"widget.1.alert_graph_definition.0.live_span = 1h",
	"widget.1.alert_graph_definition.0.alert_id = 895606",
	"widget.0.alert_graph_definition.0.title_size =",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.alert_graph_definition.0.title_align =",
	"widget.0.alert_graph_definition.0.title =",
	"widget.1.alert_graph_definition.0.title_align = right",
	"layout_type = ordered",
	"widget.0.alert_graph_definition.0.viz_type = timeseries",
	// Deprecated widget
	"widget.2.alert_graph_definition.0.time.% = 1",
	"widget.2.alert_graph_definition.0.title = Widget Title",
	"widget.2.alert_graph_definition.0.title_size = 16",
	"widget.2.alert_graph_definition.0.viz_type = toplist",
	"widget.2.alert_graph_definition.0.time.live_span = 1h",
	"widget.2.alert_graph_definition.0.alert_id = 895606",
	"widget.2.alert_graph_definition.0.title_align = right",
}

func TestAccDatadogDashboardAlertGraph(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardAlertGraphConfig, "datadog_dashboard.alert_graph_dashboard", datadogDashboardAlertGraphAsserts)
}

func TestAccDatadogDashboardAlertGraph_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardAlertGraphConfig, "datadog_dashboard.alert_graph_dashboard")
}
