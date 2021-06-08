package test

import (
	"testing"
)

const datadogDashboardHeatMapConfig = `
resource "datadog_dashboard" "heatmap_dashboard" {
	title         = "{{uniq}}"
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

			live_span = "1mo"
			event {
				q = "env:prod"
				tags_execution = "and"
			}
			show_legend = true
			legend_size = "2"
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`
const datadogDashboardHeatMapConfigImport = `
resource "datadog_dashboard" "heatmap_dashboard" {
	title         = "{{uniq}}"
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

			live_span = "1mo"
			event {
				q = "env:prod"
				tags_execution = "and"
			}
			show_legend = true
			legend_size = "2"
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`

var datadogDashboardHeatMapAsserts = []string{
	"title = {{uniq}}",
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
	"widget.0.heatmap_definition.0.live_span = 1mo",
	"widget.0.heatmap_definition.0.event.0.q = env:prod",
	"widget.0.heatmap_definition.0.event.0.tags_execution = and",
	"widget.0.heatmap_definition.0.show_legend = true",
	"widget.0.heatmap_definition.0.legend_size = 2",
	"widget.0.heatmap_definition.0.custom_link.# = 1",
	"widget.0.heatmap_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.heatmap_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

func TestAccDatadogDashboardHeatMap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardHeatMapConfig, "datadog_dashboard.heatmap_dashboard", datadogDashboardHeatMapAsserts)
}

func TestAccDatadogDashboardHeatMap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardHeatMapConfigImport, "datadog_dashboard.heatmap_dashboard")
}
