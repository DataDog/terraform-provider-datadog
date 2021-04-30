package test

import (
	"testing"
)

const datadogDashboardQueryValueConfig = `
resource "datadog_dashboard" "query_value_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		query_value_definition {
			title = "Avg of system.mem.free over account:prod"
			title_align = "center"
			title_size = "16"
			custom_unit = "Gib"
			precision = "3"
			autoscale = "true"
			request {
				q = "avg:system.mem.free{account:prod}"
				aggregator = "max"
				conditional_formats {
					palette = "white_on_red"
					value = "9"
					comparator = "<"
				}
				conditional_formats {
					palette = "white_on_green"
					value = "9"
					comparator = ">="
				}
			}
			live_span = "1h"
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`

const datadogDashboardQueryValueConfigImport = `
resource "datadog_dashboard" "query_value_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		query_value_definition {
			title = "Avg of system.mem.free over account:prod"
			title_align = "center"
			title_size = "16"
			custom_unit = "Gib"
			precision = "3"
			autoscale = "true"
			request {
				q = "avg:system.mem.free{account:prod}"
				aggregator = "max"
				conditional_formats {
					palette = "white_on_red"
					value = "9"
					comparator = "<"
				}
				conditional_formats {
					palette = "white_on_green"
					value = "9"
					comparator = ">="
				}
			}
			live_span = "1h"
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`

const datadogDashboardQueryValueFormulaConfig = `
resource "datadog_dashboard" "query_value_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		query_value_definition {
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
		}
	  }
	  widget {
		query_value_definition {
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
			  }
			}
		  }
		}
	  }
	
}
`

var datadogDashboardQueryValueAsserts = []string{
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.0.query_value_definition.0.live_span = 1h",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.palette = white_on_red",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.image_url =",
	"widget.0.query_value_definition.0.precision = 3",
	"widget.0.query_value_definition.0.request.0.aggregator = max",
	"layout_type = ordered",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.palette = white_on_green",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.custom_fg_color =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.value = 9",
	"widget.0.query_value_definition.0.autoscale = true",
	"widget.0.query_value_definition.0.request.0.q = avg:system.mem.free{account:prod}",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.comparator = >=",
	"widget.0.query_value_definition.0.title_size = 16",
	"widget.0.query_value_definition.0.custom_unit = Gib",
	"widget.0.query_value_definition.0.title_align = center",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.value = 9",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.timeframe =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.image_url =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.hide_value = false",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.timeframe =",
	"widget.0.query_value_definition.0.text_align =",
	"widget.0.query_value_definition.0.title = Avg of system.mem.free over account:prod",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.custom_bg_color =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.# = 2",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.custom_fg_color =",
	"is_read_only = true",
	"title = {{uniq}}",
	"widget.0.query_value_definition.0.custom_link.# = 1",
	"widget.0.query_value_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.query_value_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

var datadogDashboardQueryValueFormulaAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.query_value_definition.0.request.0.formula.0.formula_expression = query1 + query2",
	"widget.0.query_value_definition.0.request.0.formula.0.limit.0.count = 10",
	"widget.0.query_value_definition.0.request.0.formula.0.limit.0.order = asc",
	"widget.0.query_value_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.query_value_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.system{*} by {datacenter}",
	"widget.0.query_value_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.query_value_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.query_value_definition.0.request.0.query.1.metric_query.0.data_source = metrics",
	"widget.0.query_value_definition.0.request.0.query.1.metric_query.0.query = avg:system.load.1{*} by {datacenter}",
	"widget.0.query_value_definition.0.request.0.query.1.metric_query.0.name = query2",
	"widget.0.query_value_definition.0.request.0.query.1.metric_query.0.aggregator = sum",
	"widget.1.query_value_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.1.query_value_definition.0.request.0.formula.0.limit.0.count = 25",
	"widget.1.query_value_definition.0.request.0.formula.0.limit.0.order = desc",
	"widget.1.query_value_definition.0.request.0.query.0.event_query.0.data_source = rum",
	"widget.1.query_value_definition.0.request.0.query.0.event_query.0.indexes.# = 1",
	"widget.1.query_value_definition.0.request.0.query.0.event_query.0.indexes.0 = *",
	"widget.1.query_value_definition.0.request.0.query.0.event_query.0.name = query1",
	"widget.1.query_value_definition.0.request.0.query.0.event_query.0.compute.0.aggregation = count",
}

func TestAccDatadogDashboardQueryValue(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardQueryValueConfig, "datadog_dashboard.query_value_dashboard", datadogDashboardQueryValueAsserts)
}

func TestAccDatadogDashboardQueryValue_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardQueryValueConfigImport, "datadog_dashboard.query_value_dashboard")
}

func TestAccDatadogDashboardQueryValueFormula(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardQueryValueFormulaConfig, "datadog_dashboard.query_value_dashboard", datadogDashboardQueryValueFormulaAsserts)
}

func TestAccDatadogDashboardQueryValueFormula_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardQueryValueFormulaConfig, "datadog_dashboard.query_value_dashboard")
}
