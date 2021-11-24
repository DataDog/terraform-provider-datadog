package test

import (
	"testing"
)

const datadogDashboardChangeConfigImport = `
resource "datadog_dashboard" "change_dashboard" {
   	title         = "{{uniq}}"
   	description   = "Created using the Datadog provider in Terraform"
   	layout_type   = "ordered"
   	is_read_only  = true
	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
			}
		}
	}

	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
				compare_to = "day_before"
				increase_good = "false"
				order_by = "change"
				change_type = "absolute"
				order_dir = "desc"
				show_present = "true"
			}
			title = "Sum of system.cpu.user over * by service,account"
			title_size = "16"
			title_align = "left"
			live_span = "1h"
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`

const datadogDashboardChangeConfig = `
resource "datadog_dashboard" "change_dashboard" {
   	title         = "{{uniq}}"
   	description   = "Created using the Datadog provider in Terraform"
   	layout_type   = "ordered"
   	is_read_only  = true
	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
			}
		}
	}

	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
				compare_to = "day_before"
				increase_good = "false"
				order_by = "change"
				change_type = "absolute"
				order_dir = "desc"
				show_present = "true"
			}
			title = "Sum of system.cpu.user over * by service,account"
			title_size = "16"
			title_align = "left"
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
}
`

const datadogDashboardChangeFormulaConfig = `
resource "datadog_dashboard" "change_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		change_definition {
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
		change_definition {
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
		  custom_link {
			label = "my custom link"
			link  = "https://app.datadoghq.com/dashboard/lists"
		  }
		}
	  }
}
`

var datadogDashboardChangeAsserts = []string{
	"widget.0.change_definition.0.request.0.q = sum:system.cpu.user{*} by {service,account}",
	"widget.1.change_definition.0.title_align = left",
	"widget.1.change_definition.0.request.0.change_type = absolute",
	"widget.0.change_definition.0.request.0.order_dir =",
	"widget.0.change_definition.0.title_size =",
	"title = {{uniq}}",
	"widget.0.change_definition.0.request.0.change_type =",
	"widget.1.change_definition.0.title = Sum of system.cpu.user over * by service,account",
	"widget.1.change_definition.0.title_size = 16",
	"widget.1.change_definition.0.request.0.compare_to = day_before",
	"is_read_only = true",
	"widget.0.change_definition.0.title_align =",
	"widget.0.change_definition.0.title =",
	"widget.1.change_definition.0.request.0.q = sum:system.cpu.user{*} by {service,account}",
	"widget.1.change_definition.0.request.0.show_present = true",
	"widget.1.change_definition.0.request.0.order_by = change",
	"layout_type = ordered",
	"widget.1.change_definition.0.request.0.order_dir = desc",
	"widget.0.change_definition.0.request.0.increase_good = false",
	"widget.1.change_definition.0.request.0.increase_good = false",
	"widget.0.change_definition.0.request.0.show_present = false",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.change_definition.0.request.0.order_by =",
	"widget.1.change_definition.0.live_span = 1h",
	"widget.0.change_definition.0.request.0.compare_to =",
	"widget.1.change_definition.0.custom_link.# = 2",
	"widget.1.change_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.1.change_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.1.change_definition.0.custom_link.1.override_label = logs",
	"widget.1.change_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.1.change_definition.0.custom_link.1.is_hidden = true",
}

var datadogDashboardChangeFormulaAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.change_definition.0.request.0.formula.0.formula_expression = query1 + query2",
	"widget.0.change_definition.0.request.0.formula.0.limit.0.count = 10",
	"widget.0.change_definition.0.request.0.formula.0.limit.0.order = asc",
	"widget.0.change_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.change_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.system{*} by {datacenter}",
	"widget.0.change_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.change_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.change_definition.0.request.0.query.1.metric_query.0.data_source = metrics",
	"widget.0.change_definition.0.request.0.query.1.metric_query.0.query = avg:system.load.1{*} by {datacenter}",
	"widget.0.change_definition.0.request.0.query.1.metric_query.0.name = query2",
	"widget.0.change_definition.0.request.0.query.1.metric_query.0.aggregator = sum",
	"widget.1.change_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.1.change_definition.0.request.0.query.0.event_query.0.data_source = security_signals",
	"widget.1.change_definition.0.request.0.query.0.event_query.0.indexes.# = 1",
	"widget.1.change_definition.0.request.0.query.0.event_query.0.indexes.0 = *",
	"widget.1.change_definition.0.request.0.query.0.event_query.0.name = query1",
	"widget.1.change_definition.0.request.0.query.0.event_query.0.compute.0.aggregation = count",
}

func TestAccDatadogDashboardChange(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardChangeConfig, "datadog_dashboard.change_dashboard", datadogDashboardChangeAsserts)
}

func TestAccDatadogDashboardChange_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardChangeConfigImport, "datadog_dashboard.change_dashboard")
}

func TestAccDatadogDashboardChangeFormula(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardChangeFormulaConfig, "datadog_dashboard.change_dashboard", datadogDashboardChangeFormulaAsserts)
}

func TestAccDatadogDashboardChangeFormula_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardChangeFormulaConfig, "datadog_dashboard.change_dashboard")
}
