package test

import (
	"testing"
)

const datadogDashboardQueryTableConfig = `
resource "datadog_dashboard" "query_table_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		query_table_definition {
			title_size = "16"
			title = "system.cpu.user, system.load.1"
			title_align = "right"
			live_span = "1d"
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
}
`

const datadogDashboardQueryTableFormulaConfig = `
resource "datadog_dashboard" "query_table_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
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

const datadogDashboardQueryTableConfigImport = `
resource "datadog_dashboard" "query_table_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		query_table_definition {
			title_size = "16"
			title = "system.cpu.user, system.load.1"
			title_align = "right"
			live_span = "1d"
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
}
`

const datadogDashboardQueryTableConfigWithTextFormats = `
resource "datadog_dashboard" "query_table_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"

	widget {
		query_table_definition {
			title_size = "16"
			title = "this is hyung"
			title_align = "right"
			live_span = "1d"
			request {
				aggregator = "max"
				q = "avg:system.cpu.user{account:prod} by {service, team}"
				alias = "cpu user"
				limit = 25
				order = "desc"
				cell_display_mode = ["number"]
				text_formats {
					text_format {
						match {
							type = "is"
							value = "test"
						}
						palette = "black_on_light_yellow"
						replace {
							type = "all"
							with = "test"
				        }
					}
					text_format {
						match {
							type = "is"
							value = "versus"
						}
					}
				}
				text_formats {
					text_format {
						match {	
							type = "is"
							value = "apple"
						}
						palette = "custom_bg"
						custom_bg_color = "#632CA6"
					}
					text_format {
						match {	
							type = "is"
							value = "orange"
						}
						palette = "custom_text"
						custom_fg_color = "#632CA6"
					}
				}
			}
			has_search_bar = "auto"
		}
	}
}
`

var datadogDashboardQueryTableAsserts = []string{
	"widget.0.query_table_definition.0.live_span = 1d",
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
	"description = Created using the Datadog provider in Terraform",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.0.query_table_definition.0.request.1.q = avg:system.load.1{*} by {service, team}",
	"title = {{uniq}}",
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
	"is_read_only = true",
	"widget.0.query_table_definition.0.request.0.limit = 25",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.timeframe =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.custom_fg_color =",
	"layout_type = ordered",
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
}

var datadogDashboardQueryTableFormulaAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.query_table_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.0.query_table_definition.0.request.0.formula.0.limit.0.count = 500",
	"widget.0.query_table_definition.0.request.0.formula.0.limit.0.order = desc",
	"widget.0.query_table_definition.0.request.0.formula.0.conditional_formats.0.palette = white_on_green",
	"widget.0.query_table_definition.0.request.0.formula.0.conditional_formats.0.value = 90",
	"widget.0.query_table_definition.0.request.0.formula.0.conditional_formats.0.comparator = <",
	"widget.0.query_table_definition.0.request.0.formula.0.conditional_formats.1.palette = white_on_red",
	"widget.0.query_table_definition.0.request.0.formula.0.conditional_formats.1.value = 90",
	"widget.0.query_table_definition.0.request.0.formula.0.conditional_formats.1.comparator = >=",
	"widget.0.query_table_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.query_table_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.system{*} by {datacenter}",
	"widget.0.query_table_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.query_table_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.1.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.name = my-query",
	"widget.1.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.data_source = apm_dependency_stats",
	"widget.1.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.env = ci",
	"widget.1.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.service = cassandra",
	"widget.1.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.operation_name = cassandra.query",
	"widget.1.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.resource_name = CREATE TABLE IF NOT EXISTS foobar",
	"widget.1.query_table_definition.0.request.0.query.0.apm_dependency_stats_query.0.stat = avg_duration",
	"widget.2.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.name = my-query-2",
	"widget.2.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.data_source = apm_resource_stats",
	"widget.2.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.env = staging",
	"widget.2.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.service = foobar-controller",
	"widget.2.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.operation_name = pylons.request",
	"widget.2.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.stat = latency_p99",
	"widget.2.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.group_by.0 = resource_name",
	"widget.2.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.primary_tag_name = datacenter",
	"widget.2.query_table_definition.0.request.0.query.0.apm_resource_stats_query.0.primary_tag_value = abc",
}

var datadogDashboardQueryTableWithTextFormatsAsserts = []string{
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"widget.0.query_table_definition.0.title_size = 16",
	"widget.0.query_table_definition.0.title = this is hyung",
	"widget.0.query_table_definition.0.title_align = right",
	"widget.0.query_table_definition.0.live_span = 1d",
	"widget.0.query_table_definition.0.request.0.aggregator = max",
	"widget.0.query_table_definition.0.request.0.q = avg:system.cpu.user{account:prod} by {service, team}",
	"widget.0.query_table_definition.0.request.0.alias = cpu user",
	"widget.0.query_table_definition.0.request.0.limit = 25",
	"widget.0.query_table_definition.0.request.0.order = desc",
	"widget.0.query_table_definition.0.request.0.cell_display_mode.0 = number",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.0.match.0.type = is",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.0.match.0.value = test",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.0.palette = black_on_light_yellow",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.0.replace.0.type = all",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.0.replace.0.with = test",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.0.custom_bg_color = ",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.0.custom_fg_color = ",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.1.match.0.type = is",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.1.match.0.value = versus",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.1.palette = ",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.1.custom_bg_color = ",
	"widget.0.query_table_definition.0.request.0.text_formats.0.text_format.1.custom_fg_color = ",
	"widget.0.query_table_definition.0.request.0.text_formats.1.text_format.0.match.0.type = is",
	"widget.0.query_table_definition.0.request.0.text_formats.1.text_format.0.match.0.value = apple",
	"widget.0.query_table_definition.0.request.0.text_formats.1.text_format.0.palette = custom_bg",
	"widget.0.query_table_definition.0.request.0.text_formats.1.text_format.0.custom_bg_color = #632CA6",
	"widget.0.query_table_definition.0.request.0.text_formats.1.text_format.1.match.0.type = is",
	"widget.0.query_table_definition.0.request.0.text_formats.1.text_format.1.match.0.value = orange",
	"widget.0.query_table_definition.0.request.0.text_formats.1.text_format.1.palette = custom_text",
	"widget.0.query_table_definition.0.request.0.text_formats.1.text_format.1.custom_fg_color = #632CA6",
}

func TestAccDatadogDashboardQueryTable(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardQueryTableConfig, "datadog_dashboard.query_table_dashboard", datadogDashboardQueryTableAsserts)
}

func TestAccDatadogDashboardQueryTable_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardQueryTableConfigImport, "datadog_dashboard.query_table_dashboard")
}

func TestAccDatadogDashboardQueryTableFormula(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardQueryTableFormulaConfig, "datadog_dashboard.query_table_dashboard", datadogDashboardQueryTableFormulaAsserts)
}

func TestAccDatadogDashboardQueryTableFormula_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardQueryTableFormulaConfig, "datadog_dashboard.query_table_dashboard")
}

func TestAccDatadogDashboardQueryTableWithTextFormats(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardQueryTableConfigWithTextFormats, "datadog_dashboard.query_table_dashboard", datadogDashboardQueryTableWithTextFormatsAsserts)
}

func TestAccDatadogDashboardQueryTableWithTextFormats_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardQueryTableConfigWithTextFormats, "datadog_dashboard.query_table_dashboard")
}
