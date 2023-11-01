package test

import (
	"testing"
)

const datadogPowerpackQueryTableTest = `
resource "datadog_powerpack" "query_table_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
	widget {
		query_table_definition {
			title_size = "16"
			title = "system.cpu.user, system.load.1"
			title_align = "right"
			request {
				aggregator = "max"
				conditional_formats {
					palette = "white_on_green"
					value = 90
					comparator = "<"
				}
				conditional_formats {
					palette = "white_on_red"
					value = 90
					comparator = ">="
				}
				q = "avg:system.cpu.user{account:prod} by {service, team}"
				alias = "cpu user"
				limit = 25
				order = "desc"
				cell_display_mode = ["number"]
			}
			request {
				q = "avg:system.load.1{*} by {service, team}"
				aggregator = "last"
				conditional_formats {
					palette = "custom_bg"
					value = 50
					comparator = ">"
				}
				alias = "system load"
				cell_display_mode = ["number"]
			}
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				is_hidden = true
				override_label = "logs"
			}
			has_search_bar = "auto"
		}
	}

	widget {
		query_table_definition {
			request {
				apm_stats_query {
					service = "service"
					env = "env"
					primary_tag = "tag:*"
					name = "name"
					row_type = "resource"
				}
			}
			has_search_bar = "never"
		}
	}
	widget {
		query_table_definition {
		  request {
			formula {
			  formula_expression = "query1"
			  limit {
				count = 500
				order = "desc"
			  }
			  conditional_formats {
				palette = "white_on_green"
				value = 90
				comparator = "<"
			  }
			  conditional_formats {
			    palette = "white_on_red"
			    value = 90
				comparator = ">="
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
		  }
		}
	  }
	  widget {
		query_table_definition {
		  request {
			query {
			  apm_dependency_stats_query {
				name           = "my-query"
				data_source    = "apm_dependency_stats"
				env            = "ci"
				service        = "cassandra"
				operation_name = "cassandra.query"
				resource_name  = "CREATE TABLE IF NOT EXISTS foobar"
				stat           = "avg_duration"
			  }
			}
		  }
		}
	  }
	  widget {
		query_table_definition {
		  request {
			query {
			  apm_resource_stats_query {
				name              = "my-query-2"
				data_source       = "apm_resource_stats"
				env               = "staging"
				service           = "foobar-controller"
				operation_name    = "pylons.request"
				stat              = "latency_p99"
				group_by          = ["resource_name"]
				primary_tag_name  = "datacenter"
				primary_tag_value = "abc"
			  }
			}
		  }
		}
	  }
}
`

var datadogPowerpackQueryTableTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 5",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Query Table widget
	"widget.0.query_table_definition.0.request.1.order =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.timeframe =",
	"widget.0.query_table_definition.0.request.0.q = avg:system.cpu.user{account:prod} by {service, team}",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.custom_fg_color =",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.comparator = >",
	"widget.0.query_table_definition.0.title_size = 16",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.value = 90",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.image_url =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.hide_value = false",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.timeframe =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.0.query_table_definition.0.request.1.q = avg:system.load.1{*} by {service, team}",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.value = 90",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.custom_bg_color =",
	"widget.0.query_table_definition.0.request.1.aggregator = last",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.custom_fg_color =",
	"widget.0.query_table_definition.0.request.1.limit = 0",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.0.query_table_definition.0.request.0.aggregator = max",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.palette = custom_bg",
	"widget.0.query_table_definition.0.request.1.alias = system load",
	"widget.0.query_table_definition.0.request.0.order = desc",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.image_url =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.comparator = >=",
	"widget.0.query_table_definition.0.request.0.alias = cpu user",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.palette = white_on_red",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.value = 50",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.custom_bg_color =",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.image_url =",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.hide_value = false",
	"widget.0.query_table_definition.0.request.0.limit = 25",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.timeframe =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.custom_fg_color =",
	"widget.0.query_table_definition.0.title = system.cpu.user, system.load.1",
	"widget.0.query_table_definition.0.title_align = right",
	"widget.1.query_table_definition.0.request.0.apm_stats_query.0.service = service",
	"widget.1.query_table_definition.0.request.0.apm_stats_query.0.env = env",
	"widget.1.query_table_definition.0.request.0.apm_stats_query.0.primary_tag = tag:*",
	"widget.1.query_table_definition.0.request.0.apm_stats_query.0.name = name",
	"widget.1.query_table_definition.0.request.0.apm_stats_query.0.row_type = resource",
	"widget.0.query_table_definition.0.custom_link.# = 2",
	"widget.0.query_table_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.query_table_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.query_table_definition.0.custom_link.1.override_label = logs",
	"widget.0.query_table_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.query_table_definition.0.custom_link.1.is_hidden = true",
	"widget.0.query_table_definition.0.request.0.cell_display_mode.0 = number",
	"widget.0.query_table_definition.0.request.1.cell_display_mode.0 = number",
	"widget.0.query_table_definition.0.has_search_bar = auto",
	"widget.1.query_table_definition.0.has_search_bar = never",
	"widget.2.query_table_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.2.query_table_definition.0.request.0.formula.0.limit.0.count = 500",
	"widget.2.query_table_definition.0.request.0.formula.0.limit.0.order = desc",
	"widget.2.query_table_definition.0.request.0.formula.0.conditional_formats.0.palette = white_on_green",
	"widget.2.query_table_definition.0.request.0.formula.0.conditional_formats.0.value = 90",
	"widget.2.query_table_definition.0.request.0.formula.0.conditional_formats.0.comparator = <",
	"widget.2.query_table_definition.0.request.0.formula.0.conditional_formats.1.palette = white_on_red",
	"widget.2.query_table_definition.0.request.0.formula.0.conditional_formats.1.value = 90",
	"widget.2.query_table_definition.0.request.0.formula.0.conditional_formats.1.comparator = >=",
	"widget.2.query_table_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.2.query_table_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.system{*} by {datacenter}",
	"widget.2.query_table_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.2.query_table_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.3.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.name = my-query",
	"widget.3.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.data_source = apm_dependency_stats",
	"widget.3.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.env = ci",
	"widget.3.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.service = cassandra",
	"widget.3.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.operation_name = cassandra.query",
	"widget.3.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.resource_name = CREATE TABLE IF NOT EXISTS foobar",
	"widget.3.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.stat = avg_duration",
	"widget.4.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.name = my-query-2",
	"widget.4.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.data_source = apm_resource_stats",
	"widget.4.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.env = staging",
	"widget.4.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.service = foobar-controller",
	"widget.4.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.operation_name = pylons.request",
	"widget.4.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.stat = latency_p99",
	"widget.4.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.group_by.0 = resource_name",
	"widget.4.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.primary_tag_name = datacenter",
	"widget.4.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.primary_tag_value = abc",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackQueryTable(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackQueryTableTest, "datadog_powerpack.query_table_powerpack", datadogPowerpackQueryTableTestAsserts)
}
