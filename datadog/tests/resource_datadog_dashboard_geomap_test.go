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
		  custom_link {
			link = "https://app.datadoghq.com/dashboard/lists"
			is_hidden = true
			override_label = "logs"
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

const datadogDashboardGeomapFormulaConfig = `
resource "datadog_dashboard" "geomap_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		geomap_definition {
		  request {
			formula {
			  formula_expression = "query1 + query2"
			  limit {
				count = 10
				order = "asc"
			  }
			}
			query {
			  metric_query {
				data_source = "metrics"
				query       = "avg:system.cpu.system{*} by {datacenter}"
				name        = "query1"
				aggregator  = "sum"
			  }
			}
			query {
			  metric_query {
				data_source = "metrics"
				query       = "avg:system.load.1{*} by {datacenter}"
				name        = "query2"
				aggregator  = "sum"
			  }
			}
		  }
		  view {
			  focus = "WORLD"
		  }
		  style {
			palette      = "hostmap_blues"
			palette_flip = false
		  }
		}
	  }
	  widget {
		geomap_definition {
		  request {
			formula {
			  formula_expression = "query1"
			  limit {
				count = 25
				order = "desc"
			  }
			}
			query {
			  event_query {
				data_source = "rum"
				indexes     = ["*"]
				name        = "query1"
				compute {
				  aggregation = "count"
				}
				search {
					query = "abc"
				}
			  }
			}
		  }
		  view {
			focus = "WORLD"
		  }
		  style {
			palette      = "hostmap_blues"
			palette_flip = false
		  }
		}
	  }
	  widget {
		geomap_definition {
		  request {
			formula {
			  formula_expression = "query1"
			}
			query {
			  event_query {
				data_source = "security_signals"
				name        = "query1"
				indexes     = ["*"]
				compute {
				  aggregation = "count"
				}
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
		  custom_link {
			label = "my custom link"
			link  = "https://app.datadoghq.com/dashboard/lists"
		  }
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
	"widget.1.geomap_definition.0.custom_link.# = 2",
	"widget.1.geomap_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.1.geomap_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.1.geomap_definition.0.custom_link.1.override_label = logs",
	"widget.1.geomap_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.1.geomap_definition.0.custom_link.1.is_hidden = true",
	"widget.2.geomap_definition.0.live_span = 4h",
	"widget.2.geomap_definition.0.request.0.rum_query.0.compute_query.0.aggregation = count",
	"widget.2.geomap_definition.0.request.0.rum_query.0.index = *",
	"widget.2.geomap_definition.0.style.0.palette = hostmap_blues",
	"widget.2.geomap_definition.0.style.0.palette_flip = false",
	"widget.2.geomap_definition.0.view.0.focus = WORLD",
}

var datadogDashboardGeomapFormulaAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.geomap_definition.0.request.0.formula.0.formula_expression = query1 + query2",
	"widget.0.geomap_definition.0.request.0.formula.0.limit.0.count = 10",
	"widget.0.geomap_definition.0.request.0.formula.0.limit.0.order = asc",
	"widget.0.geomap_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.geomap_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.system{*} by {datacenter}",
	"widget.0.geomap_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.geomap_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.geomap_definition.0.request.0.query.1.metric_query.0.data_source = metrics",
	"widget.0.geomap_definition.0.request.0.query.1.metric_query.0.query = avg:system.load.1{*} by {datacenter}",
	"widget.0.geomap_definition.0.request.0.query.1.metric_query.0.name = query2",
	"widget.0.geomap_definition.0.request.0.query.1.metric_query.0.aggregator = sum",
	"widget.0.geomap_definition.0.view.0.focus = WORLD",
	"widget.0.geomap_definition.0.style.0.palette = hostmap_blues",
	"widget.0.geomap_definition.0.style.0.palette_flip = false",
	"widget.1.geomap_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.1.geomap_definition.0.request.0.formula.0.limit.0.count = 25",
	"widget.1.geomap_definition.0.request.0.formula.0.limit.0.order = desc",
	"widget.1.geomap_definition.0.request.0.query.0.event_query.0.data_source = rum",
	"widget.1.geomap_definition.0.request.0.query.0.event_query.0.indexes.# = 1",
	"widget.1.geomap_definition.0.request.0.query.0.event_query.0.indexes.0 = *",
	"widget.1.geomap_definition.0.request.0.query.0.event_query.0.name = query1",
	"widget.1.geomap_definition.0.request.0.query.0.event_query.0.search.0.query = abc",
	"widget.1.geomap_definition.0.request.0.query.0.event_query.0.compute.0.aggregation = count",
	"widget.1.geomap_definition.0.view.0.focus = WORLD",
	"widget.1.geomap_definition.0.style.0.palette_flip = false",
	"widget.1.geomap_definition.0.view.0.focus = WORLD",
	"widget.2.geomap_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.2.geomap_definition.0.request.0.query.0.event_query.0.data_source = security_signals",
	"widget.2.geomap_definition.0.request.0.query.0.event_query.0.indexes.# = 1",
	"widget.2.geomap_definition.0.request.0.query.0.event_query.0.indexes.0 = *",
	"widget.2.geomap_definition.0.request.0.query.0.event_query.0.name = query1",
	"widget.2.geomap_definition.0.request.0.query.0.event_query.0.compute.0.aggregation = count",
	"widget.2.geomap_definition.0.view.0.focus = WORLD",
	"widget.2.geomap_definition.0.style.0.palette_flip = false",
	"widget.2.geomap_definition.0.view.0.focus = WORLD",
}

func TestAccDatadogDashboardGeomap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardGeomapConfig, "datadog_dashboard.geomap_dashboard", datadogDashboardGeomapAsserts)
}

func TestAccDatadogDashboardGeomap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardGeomapConfig, "datadog_dashboard.geomap_dashboard")
}

func TestAccDatadogDashboardGeomapFormula(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardGeomapFormulaConfig, "datadog_dashboard.geomap_dashboard", datadogDashboardGeomapFormulaAsserts)
}

func TestAccDatadogDashboardGeomapFormula_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardGeomapFormulaConfig, "datadog_dashboard.geomap_dashboard")
}
