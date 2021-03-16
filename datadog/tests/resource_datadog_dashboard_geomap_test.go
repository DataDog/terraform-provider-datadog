package test

import (
	"testing"
)

const datadogDashboardGeomapConfig = `
resource "datadog_dashboard" "geomap_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		geomap_definition {
		  request {
			q = "avg:system.load.1{*}"
		  }
		  style {
			palette      = "hostmap_blues"
			palette_flip = false
		  }
		  view {
			focus = "WORLD"
		  }
		}
	  }
	  widget {
		geomap_definition {
		  request {
			log_query {
			  index = "*"
			  compute_query {
				aggregation = "count"
			  }
			}
		  }
		  style {
			palette      = "hostmap_blues"
			palette_flip = false
		  }
		  view {
			focus = "WORLD"
		  }
		  live_span = "1h"
		  custom_link {
			link = "https://app.datadoghq.com/dashboard/lists"
			label = "Test Custom Link label"
		  }
		}
	  }
	  widget {
		geomap_definition {
		  request {
			rum_query {
			  index = "*"
			  compute_query {
				aggregation = "count"
			  }
			}
		  }
		  style {
			palette      = "hostmap_blues"
			palette_flip = false
		  }
		  view {
			focus = "WORLD"
		  }
		  live_span = "4h"
		}
	  }
}
`

var datadogDashboardGeomapAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"title = {{uniq}}",
	"is_read_only = true",
	"widget.0.geomap_definition.0.request.0.q = avg:system.load.1{*}",
	"widget.0.geomap_definition.0.style.0.palette = hostmap_blues",
	"widget.0.geomap_definition.0.style.0.palette_flip = false",
	"widget.0.geomap_definition.0.view.0.focus = WORLD",
	"widget.1.geomap_definition.0.live_span = 1h",
	"widget.1.geomap_definition.0.request.0.log_query.0.compute_query.0.aggregation = count",
	"widget.1.geomap_definition.0.request.0.log_query.0.index = *",
	"widget.1.geomap_definition.0.style.0.palette = hostmap_blues",
	"widget.1.geomap_definition.0.style.0.palette_flip = false",
	"widget.1.geomap_definition.0.view.0.focus = WORLD",
	"widget.1.geomap_definition.0.custom_link.# = 1",
	"widget.1.geomap_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.1.geomap_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.2.geomap_definition.0.live_span = 4h",
	"widget.2.geomap_definition.0.request.0.rum_query.0.compute_query.0.aggregation = count",
	"widget.2.geomap_definition.0.request.0.rum_query.0.index = *",
	"widget.2.geomap_definition.0.style.0.palette = hostmap_blues",
	"widget.2.geomap_definition.0.style.0.palette_flip = false",
	"widget.2.geomap_definition.0.view.0.focus = WORLD",
}

func TestAccDatadogDashboardGeomap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardGeomapConfig, "datadog_dashboard.geomap_dashboard", datadogDashboardGeomapAsserts)
}

func TestAccDatadogDashboardGeomap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardGeomapConfig, "datadog_dashboard.geomap_dashboard")
}
