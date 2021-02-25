package test

import (
	"testing"
)

const datadogDashboardScatterplotConfig = `
resource "datadog_dashboard" "scatterplot_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		scatterplot_definition {
			title_size = "16"
			yaxis {
				scale = "log"
				include_zero = false
				min = "1"
				label = "mem (Gib)"
			}
			title_align = "right"
			color_by_groups = ["app"]
			xaxis {
				scale = "log"
				max = "100"
				min = "0"
				label = "cpu (%)"
				include_zero = false
			}
			live_span = "15m"
			title = "system.mem.used and system.cpu.user by service,team,app colored by app"
			request {
				y {
					q = "avg:system.mem.used{env:prod} by {service, team, app}"
					aggregator = "avg"
				}
				x {
					q = "avg:system.cpu.user{account:prod} by {service, team, app}"
					aggregator = "avg"
				}
			}
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
	widget {
		scatterplot_definition {
			title_size = "16"
			yaxis {
				scale = "log"
				include_zero = false
				min = "1"
				label = "mem (Gib)"
			}
			title_align = "right"
			color_by_groups = ["app"]
			xaxis {
				scale = "log"
				max = "100"
				min = "0"
				label = "cpu (%)"
				include_zero = false
			}
			time = {
				live_span = "15m"
			}
			title = "system.mem.used and system.cpu.user by service,team,app colored by app"
			request {
				y {
					q = "avg:system.mem.used{env:prod} by {service, team, app}"
					aggregator = "avg"
				}
				x {
					q = "avg:system.cpu.user{account:prod} by {service, team, app}"
					aggregator = "avg"
				}
			}
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`

const datadogDashboardScatterplotConfigImport = `
resource "datadog_dashboard" "scatterplot_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		scatterplot_definition {
			title_size = "16"
			yaxis {
				scale = "log"
				include_zero = false
				min = "1"
				label = "mem (Gib)"
			}
			title_align = "right"
			color_by_groups = ["app"]
			xaxis {
				scale = "log"
				max = "100"
				min = "0"
				label = "cpu (%)"
				include_zero = false
			}
			live_span = "15m"
			title = "system.mem.used and system.cpu.user by service,team,app colored by app"
			request {
				y {
					q = "avg:system.mem.used{env:prod} by {service, team, app}"
					aggregator = "avg"
				}
				x {
					q = "avg:system.cpu.user{account:prod} by {service, team, app}"
					aggregator = "avg"
				}
			}
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`

var datadogDashboardScatterplotAsserts = []string{
	"widget.0.scatterplot_definition.0.xaxis.0.min = 0",
	"widget.0.scatterplot_definition.0.color_by_groups.0 = app",
	"widget.0.scatterplot_definition.0.title = system.mem.used and system.cpu.user by service,team,app colored by app",
	"widget.0.scatterplot_definition.0.xaxis.0.max = 100",
	"widget.0.scatterplot_definition.0.yaxis.0.scale = log",
	"is_read_only = true",
	"widget.0.scatterplot_definition.0.title_size = 16",
	"widget.0.scatterplot_definition.0.yaxis.0.min = 1",
	"widget.0.scatterplot_definition.0.yaxis.0.label = mem (Gib)",
	"widget.0.scatterplot_definition.0.xaxis.0.include_zero = false",
	"widget.0.scatterplot_definition.0.request.0.x.0.q = avg:system.cpu.user{account:prod} by {service, team, app}",
	"widget.0.scatterplot_definition.0.title_align = right",
	"layout_type = ordered",
	"title = {{uniq}}",
	"widget.0.scatterplot_definition.0.request.0.x.0.aggregator = avg",
	"widget.0.scatterplot_definition.0.yaxis.0.include_zero = false",
	"widget.0.scatterplot_definition.0.live_span = 15m",
	"widget.0.scatterplot_definition.0.yaxis.0.max =",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.scatterplot_definition.0.request.0.y.0.q = avg:system.mem.used{env:prod} by {service, team, app}",
	"widget.0.scatterplot_definition.0.xaxis.0.label = cpu (%)",
	"widget.0.scatterplot_definition.0.request.0.y.0.aggregator = avg",
	"widget.0.scatterplot_definition.0.xaxis.0.scale = log",
	"widget.0.scatterplot_definition.0.custom_link.# = 1",
	"widget.0.scatterplot_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.scatterplot_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	// Deprecated widget
	"widget.1.scatterplot_definition.0.xaxis.0.min = 0",
	"widget.1.scatterplot_definition.0.color_by_groups.0 = app",
	"widget.1.scatterplot_definition.0.title = system.mem.used and system.cpu.user by service,team,app colored by app",
	"widget.1.scatterplot_definition.0.xaxis.0.max = 100",
	"widget.1.scatterplot_definition.0.yaxis.0.scale = log",
	"widget.1.scatterplot_definition.0.title_size = 16",
	"widget.1.scatterplot_definition.0.yaxis.0.min = 1",
	"widget.1.scatterplot_definition.0.yaxis.0.label = mem (Gib)",
	"widget.1.scatterplot_definition.0.xaxis.0.include_zero = false",
	"widget.1.scatterplot_definition.0.request.0.x.0.q = avg:system.cpu.user{account:prod} by {service, team, app}",
	"widget.1.scatterplot_definition.0.title_align = right",
	"widget.1.scatterplot_definition.0.request.0.x.0.aggregator = avg",
	"widget.1.scatterplot_definition.0.yaxis.0.include_zero = false",
	"widget.1.scatterplot_definition.0.time.live_span = 15m",
	"widget.1.scatterplot_definition.0.yaxis.0.max =",
	"widget.1.scatterplot_definition.0.request.0.y.0.q = avg:system.mem.used{env:prod} by {service, team, app}",
	"widget.1.scatterplot_definition.0.xaxis.0.label = cpu (%)",
	"widget.1.scatterplot_definition.0.request.0.y.0.aggregator = avg",
	"widget.1.scatterplot_definition.0.xaxis.0.scale = log",
	"widget.1.scatterplot_definition.0.custom_link.# = 1",
	"widget.1.scatterplot_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.1.scatterplot_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

func TestAccDatadogDashboardScatterplot(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardScatterplotConfig, "datadog_dashboard.scatterplot_dashboard", datadogDashboardScatterplotAsserts)
}

func TestAccDatadogDashboardScatterplot_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardScatterplotConfigImport, "datadog_dashboard.scatterplot_dashboard")
}
