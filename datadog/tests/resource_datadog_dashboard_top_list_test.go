package test

import (
	"testing"
)

const datadogDashboardTopListConfig = `
resource "datadog_dashboard" "top_list_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		toplist_definition {
			title_size = "16"
			title = "Avg of system.core.user over account:prod by service,app"
			title_align = "right"
			live_span = "1w"
			request {
				q = "top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')"
				conditional_formats {
					palette = "white_on_red"
					value = 15000
					comparator = ">"
				}
			}
      		custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
	widget {
		toplist_definition {
			title_size = "16"
			title = "Avg of system.core.user over account:prod by service,app"
			title_align = "right"
			time = {
				live_span = "1w"
			}
			request {
				q = "top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')"
				conditional_formats {
					palette = "white_on_red"
					value = 15000
					comparator = ">"
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

const datadogDashboardTopListConfigImport = `
resource "datadog_dashboard" "top_list_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		toplist_definition {
			title_size = "16"
			title = "Avg of system.core.user over account:prod by service,app"
			title_align = "right"
			live_span = "1w"
			request {
				q = "top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')"
				conditional_formats {
					palette = "white_on_red"
					value = 15000
					comparator = ">"
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

const datadogDashboardTopListFormulaConfig = `
resource "datadog_dashboard" "top_list_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		toplist_definition {
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
		toplist_definition {
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
		}
	  }
	
}
`

var datadogDashboardTopListAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.timeframe =",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.image_url =",
	"layout_type = ordered",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.comparator = >",
	"title = {{uniq}}",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.palette = white_on_red",
	"is_read_only = true",
	"widget.0.toplist_definition.0.live_span = 1w",
	"widget.0.toplist_definition.0.time.% = 0",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.0.toplist_definition.0.request.0.q = top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')",
	"widget.0.toplist_definition.0.title_size = 16",
	"widget.0.toplist_definition.0.title_align = right",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.value = 15000",
	"widget.0.toplist_definition.0.title = Avg of system.core.user over account:prod by service,app",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.custom_fg_color =",
	"widget.0.toplist_definition.0.custom_link.# = 1",
	"widget.0.toplist_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.toplist_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	// Deprecated widget
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.timeframe =",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.image_url =",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.comparator = >",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.palette = white_on_red",
	"widget.1.toplist_definition.0.time.live_span = 1w",
	"widget.1.toplist_definition.0.time.% = 1",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.1.toplist_definition.0.request.0.q = top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')",
	"widget.1.toplist_definition.0.title_size = 16",
	"widget.1.toplist_definition.0.title_align = right",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.value = 15000",
	"widget.1.toplist_definition.0.title = Avg of system.core.user over account:prod by service,app",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.custom_fg_color =",
	"widget.1.toplist_definition.0.custom_link.# = 1",
	"widget.1.toplist_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.1.toplist_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

var datadogDashboardTopListFormulaAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.toplist_definition.0.request.0.formula.0.formula_expression = query1 + query2",
	"widget.0.toplist_definition.0.request.0.formula.0.limit.0.count = 10",
	"widget.0.toplist_definition.0.request.0.formula.0.limit.0.order = asc",
	"widget.0.toplist_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.toplist_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.system{*} by {datacenter}",
	"widget.0.toplist_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.toplist_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.toplist_definition.0.request.0.query.1.metric_query.0.data_source = metrics",
	"widget.0.toplist_definition.0.request.0.query.1.metric_query.0.query = avg:system.load.1{*} by {datacenter}",
	"widget.0.toplist_definition.0.request.0.query.1.metric_query.0.name = query2",
	"widget.0.toplist_definition.0.request.0.query.1.metric_query.0.aggregator = sum",
	"widget.1.toplist_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.1.toplist_definition.0.request.0.formula.0.limit.0.count = 25",
	"widget.1.toplist_definition.0.request.0.formula.0.limit.0.order = desc",
	"widget.1.toplist_definition.0.request.0.query.0.event_query.0.data_source = rum",
	"widget.1.toplist_definition.0.request.0.query.0.event_query.0.indexes.# = 1",
	"widget.1.toplist_definition.0.request.0.query.0.event_query.0.indexes.0 = *",
	"widget.1.toplist_definition.0.request.0.query.0.event_query.0.name = query1",
	"widget.1.toplist_definition.0.request.0.query.0.event_query.0.search.0.query = abc",
	"widget.1.toplist_definition.0.request.0.query.0.event_query.0.compute.0.aggregation = count",
}

func TestAccDatadogDashboardTopList(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTopListConfig, "datadog_dashboard.top_list_dashboard", datadogDashboardTopListAsserts)
}

func TestAccDatadogDashboardTopList_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardTopListConfigImport, "datadog_dashboard.top_list_dashboard")
}

func TestAccDatadogDashboardTopListFormula(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTopListFormulaConfig, "datadog_dashboard.top_list_dashboard", datadogDashboardTopListFormulaAsserts)
}

func TestAccDatadogDashboardTopListFormula_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardTopListFormulaConfig, "datadog_dashboard.top_list_dashboard")
}
