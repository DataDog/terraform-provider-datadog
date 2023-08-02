package test

import (
	"testing"
)

const datadogDashboardTimeseriesConfig = `
resource "datadog_dashboard" "timeseries_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		timeseries_definition {
			title_size = "16"
			title_align = "left"
			show_legend = "true"
			title = "system.cpu.user, env, process.stat.cpu.total_pct, network.bytes_read, @d..."
			legend_size = "2"
			yaxis {
				label = ""
				min = "0"
				include_zero = "true"
				max = "599999"
				scale = ""
			}
			right_yaxis {
				label = ""
				min = "1"
				include_zero = "false"
				max = "599998"
				scale = ""
			}
			marker {
				display_type = "error dashed"
				value = "y=500000"
				label = "y=500000"
			}
			marker {
				display_type = "warning dashed"
				value = "y=400000"
				label = "y=400000"
			}
			live_span = "5m"
			event {
				q = "sources:test tags:1"
				tags_execution = "and"
			}
			request {
				q = "avg:system.cpu.user{env:prod} by {app}"
				style {
					line_width = "thin"
					palette = "dog_classic"
					line_type = "solid"
				}
				display_type = "line"
				on_right_yaxis = "true"
				metadata {
					// See https://github.com/DataDog/terraform-provider-datadog/issues/861
					expression = ""
				}
			}
			request {
				style {
					line_width = "normal"
					palette = "cool"
					line_type = "solid"
				}
				display_type = "line"
				log_query {
					index = "*"
					search_query = ""
					group_by {
						facet = "service"
						sort_query {
							aggregation = "count"
							order = "desc"
						}
						limit = "10"
					}
					compute_query {
						aggregation = "count"
					}
				}
				on_right_yaxis = "false"
			}
			request {
				style {
					line_width = "thick"
					palette = "warm"
					line_type = "dashed"
				}
				apm_query {
					index = "trace-search"
					search_query = ""
					group_by {
						facet = "status"
						sort_query {
							facet = "env"
							aggregation = "cardinality"
							order = "desc"
						}
						limit = "10"
					}
					compute_query {
						facet = "env"
						interval = 1000
						aggregation = "cardinality"
					}
				}
				display_type = "line"
				on_right_yaxis = "true"
			}
			request {
				style {
					line_width = "normal"
					palette = "purple"
					line_type = "solid"
				}
				process_query {
					search_by = ""
					metric = "process.stat.cpu.total_pct"
					limit = "10"
					filter_by = ["account:prod"]
				}
				display_type = "line"
				on_right_yaxis = "true"
			}
			request {
				style {
					line_width = "normal"
					palette = "orange"
					line_type = "solid"
				}
				display_type = "area"
				network_query {
					index = "netflow-search"
					search_query = "network.transport:udp network.destination.ip:\"*\""
					group_by {
						facet = "source_region"
					}
					group_by {
						facet = "dest_environment"
					}
					compute_query {
						facet = "network.bytes_read"
						aggregation = "sum"
					}
				}
				on_right_yaxis = "true"
			}
			request {
				style {
					line_width = "normal"
					palette = "grey"
					line_type = "solid"
				}
				rum_query {
					index = "*"
					search_query = ""
					group_by {
						facet = "service"
						sort_query {
							facet = "@duration"
							aggregation = "avg"
							order = "desc"
						}
						limit = "10"
					}
					compute_query {
						facet = "@duration"
						interval = 10
						aggregation = "avg"
					}
				}
				display_type = "area"
				on_right_yaxis = "true"
			}
			request {
				style {
					line_width = "normal"
					palette = "green"
					line_type = "solid"
				}
				audit_query {
					index = "*"
					search_query = ""
					group_by {
						facet = "@metadata.api_key.id"
						sort_query {
							aggregation = "count"
							order = "desc"
						}
						limit = "10"
					}
					compute_query {
						aggregation = "count"
					}
				}
				display_type = "line"
				on_right_yaxis = "true"
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
			legend_layout = "horizontal"
			legend_columns = ["value", "min", "max"]
		}
	}
}
`

const datadogDashboardTimeseriesConfigImport = `
resource "datadog_dashboard" "timeseries_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		timeseries_definition {
			title_size = "16"
			title_align = "left"
			show_legend = "true"
			title = "system.cpu.user, env, process.stat.cpu.total_pct, network.bytes_read, @d..."
			legend_size = "2"
			yaxis {
				label = ""
				min = "0"
				include_zero = "true"
				max = "599999"
				scale = ""
			}
			right_yaxis {
				label = ""
				min = "1"
				include_zero = "false"
				max = "599998"
				scale = ""
			}
			marker {
				display_type = "error dashed"
				value = "y=500000"
				label = "y=500000"
			}
			marker {
				display_type = "warning dashed"
				value = "y=400000"
				label = "y=400000"
			}
			live_span = "5m"
			event {
				q = "sources:test tags:1"
				tags_execution = "and"
			}
			request {
				q = "avg:system.cpu.user{env:prod} by {app}"
				style {
					line_width = "thin"
					palette = "dog_classic"
					line_type = "solid"
				}
				display_type = "line"
				on_right_yaxis = "true"
				metadata {
					// See https://github.com/DataDog/terraform-provider-datadog/issues/861
					expression = ""
				}
			}
			request {
				style {
					line_width = "normal"
					palette = "cool"
					line_type = "solid"
				}
				display_type = "line"
				log_query {
					index = "*"
					search_query = ""
					group_by {
						facet = "service"
						sort_query {
							aggregation = "count"
							order = "desc"
						}
						limit = "10"
					}
					compute_query {
						aggregation = "count"
					}
				}
				on_right_yaxis = "false"
			}
			request {
				style {
					line_width = "thick"
					palette = "warm"
					line_type = "dashed"
				}
				apm_query {
					index = "trace-search"
					search_query = ""
					group_by {
						facet = "status"
						sort_query {
							facet = "env"
							aggregation = "cardinality"
							order = "desc"
						}
						limit = "10"
					}
					compute_query {
						facet = "env"
						interval = 1000
						aggregation = "cardinality"
					}
				}
				display_type = "line"
				on_right_yaxis = "true"
			}
			request {
				style {
					line_width = "normal"
					palette = "purple"
					line_type = "solid"
				}
				process_query {
					search_by = ""
					metric = "process.stat.cpu.total_pct"
					limit = "10"
					filter_by = ["account:prod"]
				}
				display_type = "line"
				on_right_yaxis = "true"
			}
			request {
				style {
					line_width = "normal"
					palette = "orange"
					line_type = "solid"
				}
				display_type = "area"
				network_query {
					index = "netflow-search"
					search_query = "network.transport:udp network.destination.ip:\"*\""
					group_by {
						facet = "source_region"
					}
					group_by {
						facet = "dest_environment"
					}
					compute_query {
						facet = "network.bytes_read"
						aggregation = "sum"
					}
				}
				on_right_yaxis = "true"
			}
			request {
				style {
					line_width = "normal"
					palette = "grey"
					line_type = "solid"
				}
				rum_query {
					index = "*"
					search_query = ""
					group_by {
						facet = "service"
						sort_query {
							facet = "@duration"
							aggregation = "avg"
							order = "desc"
						}
						limit = "10"
					}
					compute_query {
						facet = "@duration"
						interval = 10
						aggregation = "avg"
					}
				}
				display_type = "area"
				on_right_yaxis = "true"
			}
			request {
				style {
					line_width = "normal"
					palette = "green"
					line_type = "solid"
				}
				audit_query {
					index = "*"
					search_query = ""
					group_by {
						facet = "@metadata.api_key.id"
						sort_query {
							aggregation = "count"
							order = "desc"
						}
						limit = "10"
					}
					compute_query {
						aggregation = "count"
					}
				}
				display_type = "line"
				on_right_yaxis = "true"
			}
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
			legend_layout = "horizontal"
			legend_columns = ["value", "min", "max"]
		}
	}
}
`

const datadogDashboardFormulaConfig = `
resource "datadog_dashboard" "timeseries_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		timeseries_definition {
			request {
				formula {
					formula_expression = "my_query_1 + my_query_2"
					limit {
						count = 5
						order = "asc"
					}
					alias = "sum query"
				}
				formula {
					formula_expression = "my_query_1 * my_query_2"
					limit {
						count = 7
						order = "desc"
					}
					alias = "multiplicative query"
				}
				query {
					 metric_query {
						 data_source = "metrics"
						 query = "avg:system.cpu.user{app:general} by {env}"
						 name = "my_query_1"
						 aggregator = "sum"
					}
				}
				query {
					 metric_query {
						 data_source = "metrics"
						 query = "avg:system.cpu.user{app:general} by {env}"
						 name = "my_query_2"
						 aggregator = "sum"
					}
				}
			}
		}
	}
	widget {
		timeseries_definition {
			request {
				query {
					event_query {
						data_source = "logs"
						indexes = ["days-3"]
						storage = "hot"
						name = "my_event_query"
						compute {
							aggregation = "count"
						}
						search {
							query = "abc"
						}
						group_by {
							facet = "host"
							sort {
							  metric = "@lambda.max_memory_used"
							  aggregation = "avg"
							  order = "desc"
							}
							limit = 10
						}
					}
				}
			}
			request {
				display_type = "overlay"
				query {
					event_query {
						data_source = "logs"
						name = "my_event_overlay"
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
	widget {
		timeseries_definition {
			request {
				query {
					process_query {
						data_source = "process"
						text_filter = "abc"
						metric = "process.stat.cpu.total_pct"
						limit = 10
						tag_filters = ["some_filter"]
						name = "my_process_query"
						sort = "asc"
						is_normalized_cpu = true
					}
				}
			}
		}
	}
	widget {
		timeseries_definition {
			request {
				query {
					slo_query {
						data_source = "slo"
						slo_id = "b4c7739b2af25f9d947f828730357832"
						name = "query1"
						group_mode = "overall"
						measure = "slo_status"
						slo_query_type = "metric"
						additional_query_filters = "*"
					}
				}
			}
		}
	}
}
`

var datadogDashboardTimeseriesAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.timeseries_definition.0.show_legend = true",
	"widget.0.timeseries_definition.0.yaxis.0.min = 0",
	"widget.0.timeseries_definition.0.yaxis.0.max = 599999",
	"widget.0.timeseries_definition.0.yaxis.0.label =",
	"widget.0.timeseries_definition.0.yaxis.0.include_zero = true",
	"widget.0.timeseries_definition.0.yaxis.0.scale =",
	"widget.0.timeseries_definition.0.right_yaxis.0.min = 1",
	"widget.0.timeseries_definition.0.right_yaxis.0.max = 599998",
	"widget.0.timeseries_definition.0.right_yaxis.0.label =",
	"widget.0.timeseries_definition.0.right_yaxis.0.include_zero = false",
	"widget.0.timeseries_definition.0.right_yaxis.0.scale =",
	"widget.0.timeseries_definition.0.legend_size = 2",
	"widget.0.timeseries_definition.0.live_span = 5m",
	"widget.0.timeseries_definition.0.title_align = left",
	"widget.0.timeseries_definition.0.title = system.cpu.user, env, process.stat.cpu.total_pct, network.bytes_read, @d...",
	"widget.0.timeseries_definition.0.title_size = 16",
	"widget.0.timeseries_definition.0.event.0.q = sources:test tags:1",
	"widget.0.timeseries_definition.0.event.0.tags_execution = and",
	"widget.0.timeseries_definition.0.marker.# = 2",
	"widget.0.timeseries_definition.0.marker.0.label = y=500000",
	"widget.0.timeseries_definition.0.marker.0.value = y=500000",
	"widget.0.timeseries_definition.0.marker.0.display_type = error dashed",
	"widget.0.timeseries_definition.0.marker.1.label = y=400000",
	"widget.0.timeseries_definition.0.marker.1.display_type = warning dashed",
	"widget.0.timeseries_definition.0.marker.1.value = y=400000",
	"widget.0.timeseries_definition.0.request.# = 7",
	"widget.0.timeseries_definition.0.request.0.style.0.line_width = thin",
	"widget.0.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.0.process_query.# = 0",
	"widget.0.timeseries_definition.0.request.0.metadata.# = 1",
	"widget.0.timeseries_definition.0.request.0.log_query.# = 0",
	"widget.0.timeseries_definition.0.request.0.display_type = line",
	"widget.0.timeseries_definition.0.request.0.style.# = 1",
	"widget.0.timeseries_definition.0.request.0.apm_query.# = 0",
	"widget.0.timeseries_definition.0.request.0.style.0.palette = dog_classic",
	"widget.0.timeseries_definition.0.request.0.q = avg:system.cpu.user{env:prod} by {app}",
	"widget.0.timeseries_definition.0.request.0.on_right_yaxis = true",
	"widget.0.timeseries_definition.0.request.1.log_query.0.index = *",
	"widget.0.timeseries_definition.0.request.1.style.# = 1",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.sort_query.0.aggregation = count",
	"widget.0.timeseries_definition.0.request.1.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.1.log_query.0.search_query =",
	"widget.0.timeseries_definition.0.request.1.style.0.palette = cool",
	"widget.0.timeseries_definition.0.request.1.log_query.0.compute.% = 0",
	"widget.0.timeseries_definition.0.request.1.log_query.0.compute_query.# = 1",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.facet = service",
	"widget.0.timeseries_definition.0.request.1.log_query.0.compute_query.0.aggregation = count",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.sort_query.0.order = desc",
	"widget.0.timeseries_definition.0.request.1.metadata.# = 0",
	"widget.0.timeseries_definition.0.request.1.q =",
	"widget.0.timeseries_definition.0.request.1.apm_query.# = 0",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.# = 1",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.1.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.1.log_query.0.group_by.0.sort_query.# = 1",
	"widget.0.timeseries_definition.0.request.1.process_query.# = 0",
	"widget.0.timeseries_definition.0.request.1.display_type = line",
	"widget.0.timeseries_definition.0.request.1.log_query.# = 1",
	"widget.0.timeseries_definition.0.request.1.on_right_yaxis = false",
	"widget.0.timeseries_definition.0.request.3.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.3.process_query.0.metric = process.stat.cpu.total_pct",
	"widget.0.timeseries_definition.0.request.2.style.0.line_type = dashed",
	"widget.0.timeseries_definition.0.request.2.display_type = line",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.facet = status",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.# = 1",
	"widget.0.timeseries_definition.0.request.2.apm_query.# = 1",
	"widget.0.timeseries_definition.0.request.2.process_query.# = 0",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort_query.0.order = desc",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.search_query =",
	"widget.0.timeseries_definition.0.request.2.log_query.# = 0",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.compute_query.0.interval = 1000",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.compute.% = 0",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.compute_query.# = 1",
	"widget.0.timeseries_definition.0.request.2.metadata.# = 0",
	"widget.0.timeseries_definition.0.request.2.style.0.line_width = thick",
	"widget.0.timeseries_definition.0.request.2.q =",
	"widget.0.timeseries_definition.0.request.2.style.0.palette = warm",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort_query.# = 1",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.compute_query.0.facet = env",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.2.style.# = 1",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort_query.0.aggregation = cardinality",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.compute_query.0.aggregation = cardinality",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort_query.0.facet = env",
	"widget.0.timeseries_definition.0.request.2.apm_query.0.index = trace-search",
	"widget.0.timeseries_definition.0.request.2.on_right_yaxis = true",
	"widget.0.timeseries_definition.0.request.3.log_query.# = 0",
	"widget.0.timeseries_definition.0.request.3.process_query.0.search_by =",
	"widget.0.timeseries_definition.0.request.3.style.# = 1",
	"widget.0.timeseries_definition.0.request.3.metadata.# = 0",
	"widget.0.timeseries_definition.0.request.3.process_query.0.limit = 10",
	"widget.0.timeseries_definition.0.request.3.process_query.# = 1",
	"widget.0.timeseries_definition.0.request.3.process_query.0.filter_by.0 = account:prod",
	"widget.0.timeseries_definition.0.request.3.process_query.0.filter_by.# = 1",
	"widget.0.timeseries_definition.0.request.3.q =",
	"widget.0.timeseries_definition.0.request.3.display_type = line",
	"widget.0.timeseries_definition.0.request.3.apm_query.# = 0",
	"widget.0.timeseries_definition.0.request.3.style.0.palette = purple",
	"widget.0.timeseries_definition.0.request.3.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.3.on_right_yaxis = true",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.sort_query.# = 1",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.0.facet = source_region",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.1.sort_query.# = 0",
	"widget.0.timeseries_definition.0.request.4.network_query.0.compute.% = 0",
	"widget.0.timeseries_definition.0.request.4.network_query.0.compute_query.# = 1",
	"widget.0.timeseries_definition.0.request.4.network_query.0.compute_query.0.facet = network.bytes_read",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.search_query =",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.1.limit = 0",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.compute_query.0.facet = @duration",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.1.facet = dest_environment",
	"widget.0.timeseries_definition.0.request.4.network_query.0.search_query = network.transport:udp network.destination.ip:\"*\"",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.0.limit = 0",
	"widget.0.timeseries_definition.0.request.5.display_type = area",
	"widget.0.timeseries_definition.0.request.4.network_query.0.index = netflow-search",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.sort_query.0.facet = @duration",
	"widget.0.timeseries_definition.0.request.4.q =",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.compute_query.# = 1",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.compute.% = 0",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.sort_query.0.aggregation = avg",
	"widget.0.timeseries_definition.0.request.5.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.facet = service",
	"widget.0.timeseries_definition.0.request.4.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.compute_query.0.interval = 10",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.compute_query.0.aggregation = avg",
	"widget.0.timeseries_definition.0.request.5.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.4.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.4.style.0.palette = orange",
	"widget.0.timeseries_definition.0.request.4.display_type = area",
	"widget.0.timeseries_definition.0.request.4.network_query.0.group_by.0.sort_query.# = 0",
	"widget.0.timeseries_definition.0.request.5.style.0.palette = grey",
	"widget.0.timeseries_definition.0.request.4.network_query.0.compute_query.0.aggregation = sum",
	"widget.0.timeseries_definition.0.request.4.on_right_yaxis = true",
	"widget.0.timeseries_definition.0.request.5.q =",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.index = *",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.group_by.0.sort_query.0.order = desc",
	"widget.0.timeseries_definition.0.request.5.rum_query.0.search_query =",
	"widget.0.timeseries_definition.0.request.5.on_right_yaxis = true",
	"widget.0.timeseries_definition.0.request.6.audit_query.0.index = *",
	"widget.0.timeseries_definition.0.request.6.audit_query.0.compute_query.0.aggregation = count",
	"widget.0.timeseries_definition.0.request.6.audit_query.0.search_query =",
	"widget.0.timeseries_definition.0.request.6.audit_query.0.group_by.0.facet = @metadata.api_key.id",
	"widget.0.timeseries_definition.0.request.6.audit_query.0.group_by.0.sort_query.0.aggregation = count",
	"widget.0.timeseries_definition.0.request.6.audit_query.0.group_by.0.sort_query.0.order = desc",
	"widget.0.timeseries_definition.0.request.6.audit_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.6.display_type = line",
	"widget.0.timeseries_definition.0.request.6.on_right_yaxis = true",
	"widget.0.timeseries_definition.0.request.6.style.0.palette = green",
	"widget.0.timeseries_definition.0.request.6.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.6.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.legend_layout = horizontal",
	"widget.0.timeseries_definition.0.legend_columns.# = 3",
	"widget.0.timeseries_definition.0.legend_columns.TypeSet = value",
	"widget.0.timeseries_definition.0.legend_columns.TypeSet = min",
	"widget.0.timeseries_definition.0.legend_columns.TypeSet = max",
	"widget.0.timeseries_definition.0.custom_link.# = 2",
	"widget.0.timeseries_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.timeseries_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.timeseries_definition.0.custom_link.1.override_label = logs",
	"widget.0.timeseries_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.timeseries_definition.0.custom_link.1.is_hidden = true",
}

var datadogDashboardFormulaAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.timeseries_definition.0.request.0.formula.0.formula_expression = my_query_1 + my_query_2",
	"widget.0.timeseries_definition.0.request.0.formula.0.limit.0.count = 5",
	"widget.0.timeseries_definition.0.request.0.formula.0.limit.0.order = asc",
	"widget.0.timeseries_definition.0.request.0.formula.0.alias = sum query",
	"widget.0.timeseries_definition.0.request.0.formula.1.formula_expression = my_query_1 * my_query_2",
	"widget.0.timeseries_definition.0.request.0.formula.1.limit.0.count = 7",
	"widget.0.timeseries_definition.0.request.0.formula.1.limit.0.order = desc",
	"widget.0.timeseries_definition.0.request.0.formula.1.alias = multiplicative query",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{app:general} by {env}",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.name = my_query_1",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.timeseries_definition.0.request.0.query.1.metric_query.0.data_source = metrics",
	"widget.0.timeseries_definition.0.request.0.query.1.metric_query.0.query = avg:system.cpu.user{app:general} by {env}",
	"widget.0.timeseries_definition.0.request.0.query.1.metric_query.0.name = my_query_2",
	"widget.0.timeseries_definition.0.request.0.query.1.metric_query.0.aggregator = sum",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.data_source = logs",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.indexes.# = 1",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.storage = hot",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.indexes.0 = days-3",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.name = my_event_query",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.group_by.0.facet = host",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.group_by.0.sort.0.metric = @lambda.max_memory_used",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.group_by.0.sort.0.aggregation = avg",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.group_by.0.sort.0.order = desc",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.group_by.0.limit = 10",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.compute.0.aggregation = count",
	"widget.1.timeseries_definition.0.request.1.display_type = overlay",
	"widget.1.timeseries_definition.0.request.1.query.0.event_query.0.name = my_event_overlay",
	"widget.1.timeseries_definition.0.request.1.query.0.event_query.0.search.0.query = abc",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.data_source = process",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.text_filter = abc",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.metric = process.stat.cpu.total_pct",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.limit = 10",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.tag_filters.# = 1",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.tag_filters.0 = some_filter",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.name = my_process_query",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.sort = asc",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.is_normalized_cpu = true",
	"widget.3.timeseries_definition.0.request.0.query.0.slo_query.0.data_source = slo",
	"widget.3.timeseries_definition.0.request.0.query.0.slo_query.0.slo_id = b4c7739b2af25f9d947f828730357832",
	"widget.3.timeseries_definition.0.request.0.query.0.slo_query.0.name = query1",
	"widget.3.timeseries_definition.0.request.0.query.0.slo_query.0.group_mode = overall",
	"widget.3.timeseries_definition.0.request.0.query.0.slo_query.0.measure = slo_status",
	"widget.3.timeseries_definition.0.request.0.query.0.slo_query.0.slo_query_type = metric",
	"widget.3.timeseries_definition.0.request.0.query.0.slo_query.0.additional_query_filters = *",
}

func TestAccDatadogDashboardTimeseries(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTimeseriesConfig, "datadog_dashboard.timeseries_dashboard", datadogDashboardTimeseriesAsserts)
}

func TestAccDatadogDashboardTimeseries_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardTimeseriesConfigImport, "datadog_dashboard.timeseries_dashboard")
}

func TestAccDatadogDashboardFormula(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardFormulaConfig, "datadog_dashboard.timeseries_dashboard", datadogDashboardFormulaAsserts)
}

func TestAccDatadogDashboardFormula_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardFormulaConfig, "datadog_dashboard.timeseries_dashboard")
}

const datadogDashboardTimeseriesMultiComputeConfig = `
resource "datadog_dashboard" "timeseries_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		timeseries_definition {
			title_size = "16"
			title_align = "left"
			show_legend = "true"
			title = "system.cpu.user, env, process.stat.cpu.total_pct, network.bytes_read, @d..."
			legend_size = "2"
			yaxis {
				label = ""
				min = "0"
				include_zero = "true"
				max = "599999"
				scale = ""
			}
			right_yaxis {
				label = ""
				min = "1"
				include_zero = "false"
				max = "599998"
				scale = ""
			}
			marker {
				display_type = "error dashed"
				value = "y=500000"
				label = "y=500000"
			}
			marker {
				display_type = "warning dashed"
				value = "y=400000"
				label = "y=400000"
			}
			live_span = "5m"
			event {
				q = "sources:test tags:1"
				tags_execution = "and"
			}
			request {
				style {
					line_width = "normal"
					palette = "cool"
					line_type = "solid"
				}
				display_type = "line"
				log_query {
					index = "*"
					search_query = ""
					group_by {
						facet = "service"
						sort_query {
							aggregation = "count"
							order = "desc"
						}
						limit = "10"
					}
					multi_compute {
						aggregation = "count"
					}
					multi_compute {
						facet = "env"
						interval = "1000"
						aggregation = "cardinality"
					}
				}
				on_right_yaxis = "true"
			}
		}
	}
}
`

const datadogDashboardTimeseriesMultiComputeConfigImport = `
resource "datadog_dashboard" "timeseries_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	widget {
		timeseries_definition {
			title_size = "16"
			title_align = "left"
			show_legend = "true"
			title = "system.cpu.user, env, process.stat.cpu.total_pct, network.bytes_read, @d..."
			legend_size = "2"
			yaxis {
				label = ""
				min = "0"
				include_zero = "true"
				max = "599999"
				scale = ""
			}
			right_yaxis {
				label = ""
				min = "1"
				include_zero = "false"
				max = "599998"
				scale = ""
			}
			marker {
				display_type = "error dashed"
				value = "y=500000"
				label = "y=500000"
			}
			marker {
				display_type = "warning dashed"
				value = "y=400000"
				label = "y=400000"
			}
			live_span = "5m"
			event {
				q = "sources:test tags:1"
				tags_execution = "and"
			}
			request {
				style {
					line_width = "normal"
					palette = "cool"
					line_type = "solid"
				}
				display_type = "line"
				log_query {
					index = "*"
					search_query = ""
					group_by {
						facet = "service"
						sort_query {
							aggregation = "count"
							order = "desc"
						}
						limit = "10"
					}
					multi_compute {
						aggregation = "count"
					}
					multi_compute {
						facet = "env"
						interval = "1000"
						aggregation = "cardinality"
					}
				}
				on_right_yaxis = "true"
			}
		}
	}
}
`

var datadogDashboardTimeseriesMultiComputeAsserts = []string{
	"title = {{uniq}}",
	"is_read_only = true",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.timeseries_definition.0.event.0.q = sources:test tags:1",
	"widget.0.timeseries_definition.0.event.0.tags_execution = and",
	"widget.0.timeseries_definition.0.legend_size = 2",
	"widget.0.timeseries_definition.0.marker.# = 2",
	"widget.0.timeseries_definition.0.marker.0.display_type = error dashed",
	"widget.0.timeseries_definition.0.marker.0.label = y=500000",
	"widget.0.timeseries_definition.0.marker.0.value = y=500000",
	"widget.0.timeseries_definition.0.marker.1.display_type = warning dashed",
	"widget.0.timeseries_definition.0.marker.1.label = y=400000",
	"widget.0.timeseries_definition.0.marker.1.value = y=400000",
	"widget.0.timeseries_definition.0.request.# = 1",
	"widget.0.timeseries_definition.0.request.0.display_type = line",
	"widget.0.timeseries_definition.0.request.0.log_query.# = 1",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.# = 2",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.0.aggregation = count",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.1.aggregation = cardinality",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.1.facet = env",
	"widget.0.timeseries_definition.0.request.0.log_query.0.multi_compute.1.interval = 1000",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.# = 1",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.0.facet = service",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.0.sort_query.0.aggregation = count",
	"widget.0.timeseries_definition.0.request.0.log_query.0.group_by.0.sort_query.0.order = desc",
	"widget.0.timeseries_definition.0.request.0.log_query.0.index = *",
	"widget.0.timeseries_definition.0.request.0.style.# = 1",
	"widget.0.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.0.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.0.style.0.palette = cool",
	"widget.0.timeseries_definition.0.request.0.on_right_yaxis = true",
	"widget.0.timeseries_definition.0.show_legend = true",
	"widget.0.timeseries_definition.0.live_span = 5m",
	"widget.0.timeseries_definition.0.title = system.cpu.user, env, process.stat.cpu.total_pct, network.bytes_read, @d...",
	"widget.0.timeseries_definition.0.title_align = left",
	"widget.0.timeseries_definition.0.title_size = 16",
	"widget.0.timeseries_definition.0.yaxis.# = 1",
	"widget.0.timeseries_definition.0.yaxis.0.include_zero = true",
	"widget.0.timeseries_definition.0.yaxis.0.max = 599999",
	"widget.0.timeseries_definition.0.yaxis.0.min = 0",
	"widget.0.timeseries_definition.0.right_yaxis.# = 1",
	"widget.0.timeseries_definition.0.right_yaxis.0.include_zero = false",
	"widget.0.timeseries_definition.0.right_yaxis.0.max = 599998",
	"widget.0.timeseries_definition.0.right_yaxis.0.min = 1",
}

func TestAccDatadogDashboardTimeseriesMultiCompute(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTimeseriesMultiComputeConfig, "datadog_dashboard.timeseries_dashboard", datadogDashboardTimeseriesMultiComputeAsserts)
}

func TestAccDatadogDashboardTimeseriesMultiCompute_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardTimeseriesMultiComputeConfigImport, "datadog_dashboard.timeseries_dashboard")
}
