package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func datadogOrderedDashboardConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "ordered_dashboard" {
	title         = "%s"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = true
	widget {
		alert_graph_definition {
			alert_id = "895605"
			viz_type = "timeseries"
			title = "Widget Title"
			time = {
				live_span = "1h"
			}
		}
	}
	widget {
		alert_value_definition {
			alert_id = "895605"
			precision = 3
			unit = "b"
			text_align = "center"
			title = "Widget Title"
		}
	}
	widget {
		change_definition {
			request {
				q = "avg:system.load.1{env:staging} by {account}"
				change_type = "absolute"
				compare_to = "week_before"
				increase_good = true
				order_by = "name"
				order_dir = "desc"
				show_present = true
			}
			title = "Widget Title"
			time = {
				live_span = "1h"
			}
		}
	}
	widget {
		distribution_definition {
			request {
				q = "avg:system.load.1{env:staging} by {account}"
				style {
					palette = "warm"
				}
			}
			title = "Widget Title"
			time = {
				live_span = "1h"
			}
		}
	}
	widget {
		check_status_definition {
			check = "aws.ecs.agent_connected"
			grouping = "cluster"
			group_by = ["account", "cluster"]
			tags = ["account:demo", "cluster:awseb-ruthebdog-env-8-dn3m6u3gvk"]
			title = "Widget Title"
			time = {
				live_span = "1h"
			}
		}
	}
	widget {
		heatmap_definition {
			request {
				q = "avg:system.load.1{env:staging} by {account}"
				style {
					palette = "warm"
				}
			}
			yaxis {
				min = 1
				max = 2
				include_zero = true
				scale = "sqrt"
			}
			title = "Widget Title"
			time = {
				live_span = "1h"
			}
		}
	}
	widget {
		hostmap_definition {
			request {
				fill {
					q = "avg:system.load.1{*} by {host}"
				}
				size {
					q = "avg:memcache.uptime{*} by {host}"
				}
			}
			node_type= "container"
			group = ["host", "region"]
			no_group_hosts = true
			no_metric_hosts = true
			scope = ["region:us-east-1", "aws_account:727006795293"]
			style {
				palette = "yellow_to_green"
				palette_flip = true
				fill_min = "10"
				fill_max = "20"
			}
			title = "Widget Title"
		}
	}
	widget {
		note_definition {
			content = "note text"
			background_color = "pink"
			font_size = "14"
			text_align = "center"
			show_tick = true
			tick_edge = "left"
			tick_pos = "50%%" # string escaped as this is used as a format string
		}
	}
	widget {
		query_value_definition {
			request {
				q = "avg:system.load.1{env:staging} by {account}"
				aggregator = "sum"
				conditional_formats {
					comparator = "<"
					value = "2"
					palette = "white_on_green"
					metric = "system.load.1"
				}
				conditional_formats {
					comparator = ">"
					value = "2.2"
					palette = "white_on_red"
					metric = "system.load.1"
				}
			}
			autoscale = true
			custom_unit = "xx"
			precision = "4"
			text_align = "right"
			title = "Widget Title"
			time = {
				live_span = "1h"
			}
		}
	}
	widget {
		scatterplot_definition {
			request {
				x {
					q = "avg:system.cpu.user{*} by {service, account}"
					aggregator = "max"
				}
				y {
					q = "avg:system.mem.used{*} by {service, account}"
					aggregator = "min"
				}
			}
			color_by_groups = ["account", "apm-role-group"]
			xaxis {
				include_zero = true
				label = "x"
				min = "1"
				max = "2000"
				scale = "pow"
			}
			yaxis {
				include_zero = false
				label = "y"
				min = "5"
				max = "2222"
				scale = "log"
			}
			title = "Widget Title"
			time = {
				live_span = "1h"
			}
		}
	}
	widget {
		timeseries_definition {
			request {
				q = "avg:system.cpu.user{app:general} by {env}"
				display_type = "line"
				style {
					palette = "warm"
					line_type = "dashed"
					line_width = "thin"
				}
				metadata {
					expression = "avg:system.cpu.user{app:general} by {env}"
					alias_name = "Alpha"
				}
			}
			request {
				log_query {
					index = "mcnulty"
					compute = {
						aggregation = "count"
						facet = "@duration"
						interval = 5000
					}
					search = {
						query = "status:info"
					}
					group_by {
						facet = "host"
						limit = 10
						sort = {
							aggregation = "avg"
							order = "desc"
							facet = "@duration"
						}
					}
				}
				display_type = "area"
			}
			request {
				apm_query {
					index = "apm-search"
					compute = {
						aggregation = "count"
						facet = "@duration"
						interval = 5000
					}
					search = {
						query = "type:web"
					}
					group_by {
						facet = "resource_name"
						limit = 50
						sort = {
							aggregation = "avg"
							order = "desc"
							facet = "@string_query.interval"
						}
					}
				}
				display_type = "bars"
			}
			request {
				process_query {
					metric = "process.stat.cpu.total_pct"
					search_by = "error"
					filter_by = ["active"]
					limit = 50
				}
				display_type = "area"
			}
			request {
				security_query {
					index = "signal"
					compute = {
						aggregation = "count"
					}
					search = {
						query = "status:(high OR critical)"
					}
					group_by {
						facet = "status"
					}
				}
				display_type = "bars"
			}
			request {
				rum_query {
					index = "rum"
					compute = {
						aggregation = "count"
					}
					search = {
						query = "status:info"
					}
					group_by {
						facet = "service"
					}
				}
				display_type = "bars"
			}
			marker {
				display_type = "error dashed"
				label = " z=6 "
				value = "y=4"
			}
			marker {
				display_type = "ok solid"
				value = "10 < y < 999"
				label = " x=8 "
			}
			title = "Widget Title"
			show_legend = true
			legend_size = "2"
			time = {
				live_span = "1h"
			}
			event {
				q = "sources:test tags:1"
			}
			event {
				q = "sources:test tags:2"
			}
			yaxis {
				scale = "log"
				include_zero = false
				max = 100
			}
		}
	}
	widget {
		toplist_definition {
			request {
				q= "avg:system.cpu.user{app:general} by {env}"
				conditional_formats {
					comparator = "<"
					value = "2"
					palette = "white_on_green"
				}
				conditional_formats {
					comparator = ">"
					value = "2.2"
					palette = "white_on_red"
				}
			}
			title = "Widget Title"
		}
	}
	widget {
		group_definition {
			layout_type = "ordered"
			title = "Group Widget"

			widget {
				note_definition {
					content = "cluster note widget"
					background_color = "yellow"
					font_size = "16"
					text_align = "left"
					show_tick = false
					tick_edge = "left"
					tick_pos = "50%%" # string escaped as this is used as a format string
				}
			}
			widget {
				alert_graph_definition {
					alert_id = "123"
					viz_type = "toplist"
					title = "Alert Graph"
					time = {
						live_span = "1h"
					}
				}
			}
		}
	}
	widget {
		service_level_objective_definition {
			title = "Widget Title"
			view_type = "detail"
			slo_id = "56789"
			show_error_budget = true
			view_mode = "overall"
			time_windows = ["7d", "previous_week"]
		}
	}
	widget {
		query_table_definition {
			request {
				q = "avg:system.load.1{env:staging} by {account}"
				aggregator = "sum"
				limit = "10"
				conditional_formats {
					comparator = "<"
					value = "2"
					palette = "white_on_green"
				}
				conditional_formats {
					comparator = ">"
					value = "2.2"
					palette = "white_on_red"
				}
			}
			title = "Widget Title"
			time = {
				live_span = "1h"
			}
		}
	}
	template_variable {
		name   = "var_1"
		prefix = "host"
		default = "aws"
	}
	template_variable {
		name   = "var_2"
		prefix = "service_name"
		default = "autoscaling"
	}
	template_variable_preset {
		name = "preset_1"

		template_variable {
			name = "var_1"
			value = "var_1_value"
		}
		template_variable {
			name = "var_2"
			value = "var_2_value"
		}
	}
	template_variable_preset {
		name = "preset_2"

		template_variable {
			name = "var_1"
			value = "var_1_value"
		}
	}
}`, uniq)
}

func datadogFreeDashboardConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "free_dashboard" {
	title         = "%s"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = false
	widget {
		event_stream_definition {
			query = "*"
			event_size = "l"
			title = "Widget Title"
			title_size = 16
			title_align = "left"
			time = {
				live_span = "1h"
			}
		}
		layout = {
			height = 43
			width = 32
			x = 5
			y = 5
		}
	}
	widget {
		event_timeline_definition {
			query = "*"
			title = "Widget Title"
			title_size = 16
			title_align = "left"
			time = {
				live_span = "1h"
			}
		}
		layout = {
			height = 9
			width = 65
			x = 42
			y = 73
		}
	}
	widget {
		free_text_definition {
			text = "free text content"
			color = "#d00"
			font_size = "88"
			text_align = "left"
		}
		layout = {
			height = 20
			width = 30
			x = 42
			y = 5
		}
	}
	widget {
		iframe_definition {
			url = "http://google.com"
		}
		layout = {
			height = 46
			width = 39
			x = 111
			y = 8
		}
	}
	widget {
		image_definition {
			url = "https://images.pexels.com/photos/67636/rose-blue-flower-rose-blooms-67636.jpeg?auto=compress&cs=tinysrgb&h=350"
			sizing = "fit"
			margin = "small"
		}
		layout = {
			height = 20
			width = 30
			x = 77
			y = 7
		}
	}
	widget {
		log_stream_definition {
			logset = "19"
			query = "error"
			columns = ["core_host", "core_service", "tag_source"]
			show_date_column = true
			show_message_column = true
			message_display = "expanded-md"
			sort {
				column = "time"
				order = "desc"
			}
		}
		layout = {
			height = 36
			width = 32
			x = 5
			y = 51
		}
	}
	widget {
		manage_status_definition {
			color_preference = "text"
			count = 50
			display_format = "countsAndList"
			hide_zero_counts = true
			query = "type:metric"
			show_last_triggered = true
			sort = "status,asc"
			start = 0
			summary_type = "monitors"
			title = "Widget Title"
			title_size = 16
			title_align = "left"
		}
		layout = {
			height = 40
			width = 30
			x = 112
			y = 55
		}
	}
	widget {
		trace_service_definition {
			display_format = "three_column"
			env = "datad0g.com"
			service = "alerting-cassandra"
			show_breakdown = true
			show_distribution = true
			show_errors = true
			show_hits = true
			show_latency = false
			show_resource_list = false
			size_format = "large"
			span_name = "cassandra.query"
			title = "alerting-cassandra #env:datad0g.com"
			title_align = "center"
			title_size = "13"
			time = {
				live_span = "1h"
			}
		}
		layout = {
			height = 38
			width = 67
			x = 40
			y = 28
		}
	}
	template_variable {
		name   = "var_1"
		prefix = "host"
		default = "aws"
	}
	template_variable {
		name   = "var_2"
		prefix = "service_name"
		default = "autoscaling"
	}
	template_variable_preset {
		name = "preset_1"

		template_variable {
			name = "var_1"
			value = "var_1_value"
		}
		template_variable {
			name = "var_2"
			value = "var_2_value"
		}
	}
	template_variable_preset {
		name = "preset_2"

		template_variable {
			name = "var_1"
			value = "var_1_value"
		}
	}
}`, uniq)
}

var datadogOrderedDashboardAsserts = []string{
	// Dashboard metadata
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"is_read_only = true",
	"widget.# = 15",
	// Alert Graph widget
	"widget.0.alert_graph_definition.0.alert_id = 895605",
	"widget.0.alert_graph_definition.0.viz_type = timeseries",
	"widget.0.alert_graph_definition.0.title = Widget Title",
	"widget.0.alert_graph_definition.0.time.live_span = 1h",
	// Alert Value widget
	"widget.1.alert_value_definition.0.alert_id = 895605",
	"widget.1.alert_value_definition.0.precision = 3",
	"widget.1.alert_value_definition.0.unit = b",
	"widget.1.alert_value_definition.0.text_align = center",
	"widget.1.alert_value_definition.0.title = Widget Title",
	// Change widget
	"widget.2.change_definition.0.request.0.q = avg:system.load.1{env:staging} by {account}",
	"widget.2.change_definition.0.request.0.change_type = absolute",
	"widget.2.change_definition.0.request.0.compare_to = week_before",
	"widget.2.change_definition.0.request.0.increase_good = true",
	"widget.2.change_definition.0.request.0.order_by = name",
	"widget.2.change_definition.0.request.0.order_dir = desc",
	"widget.2.change_definition.0.request.0.show_present = true",
	"widget.2.change_definition.0.title = Widget Title",
	"widget.2.change_definition.0.time.live_span = 1h",
	// Distribution widget
	"widget.3.distribution_definition.0.request.0.q = avg:system.load.1{env:staging} by {account}",
	"widget.3.distribution_definition.0.request.0.style.0.palette = warm",
	"widget.3.distribution_definition.0.title = Widget Title",
	"widget.3.distribution_definition.0.time.live_span = 1h",
	// Check Status widget
	"widget.4.check_status_definition.0.check = aws.ecs.agent_connected",
	"widget.4.check_status_definition.0.grouping = cluster",
	"widget.4.check_status_definition.0.group_by.# = 2",
	"widget.4.check_status_definition.0.group_by.0 = account",
	"widget.4.check_status_definition.0.group_by.1 = cluster",
	"widget.4.check_status_definition.0.tags.# = 2",
	"widget.4.check_status_definition.0.tags.0 = account:demo",
	"widget.4.check_status_definition.0.tags.1 = cluster:awseb-ruthebdog-env-8-dn3m6u3gvk",
	"widget.4.check_status_definition.0.title = Widget Title",
	"widget.4.check_status_definition.0.time.live_span = 1h",
	// Heatmap widget
	"widget.5.heatmap_definition.0.request.0.q = avg:system.load.1{env:staging} by {account}",
	"widget.5.heatmap_definition.0.request.0.style.0.palette = warm",
	"widget.5.heatmap_definition.0.yaxis.0.min = 1",
	"widget.5.heatmap_definition.0.yaxis.0.max = 2",
	"widget.5.heatmap_definition.0.yaxis.0.include_zero = true",
	"widget.5.heatmap_definition.0.yaxis.0.scale = sqrt",
	"widget.5.heatmap_definition.0.title = Widget Title",
	"widget.5.heatmap_definition.0.time.live_span = 1h",
	// Hostmap widget
	"widget.6.hostmap_definition.0.request.0.fill.0.q = avg:system.load.1{*} by {host}",
	"widget.6.hostmap_definition.0.request.0.size.0.q = avg:memcache.uptime{*} by {host}",
	"widget.6.hostmap_definition.0.node_type = container",
	"widget.6.hostmap_definition.0.group.# = 2",
	"widget.6.hostmap_definition.0.group.0 = host",
	"widget.6.hostmap_definition.0.group.1 = region",
	"widget.6.hostmap_definition.0.scope.# = 2",
	"widget.6.hostmap_definition.0.scope.0 = region:us-east-1",
	"widget.6.hostmap_definition.0.scope.1 = aws_account:727006795293",
	"widget.6.hostmap_definition.0.style.0.palette = yellow_to_green",
	"widget.6.hostmap_definition.0.style.0.palette_flip = true",
	"widget.6.hostmap_definition.0.style.0.fill_min = 10",
	"widget.6.hostmap_definition.0.style.0.fill_max = 20",
	"widget.6.hostmap_definition.0.title = Widget Title",
	// Note widget
	"widget.7.note_definition.0.content = note text",
	"widget.7.note_definition.0.background_color = pink",
	"widget.7.note_definition.0.font_size = 14",
	"widget.7.note_definition.0.text_align = center",
	"widget.7.note_definition.0.show_tick = true",
	"widget.7.note_definition.0.tick_edge = left",
	"widget.7.note_definition.0.tick_pos = 50%",
	// Query Value widget
	"widget.8.query_value_definition.0.request.0.q = avg:system.load.1{env:staging} by {account}",
	"widget.8.query_value_definition.0.request.0.aggregator = sum",
	"widget.8.query_value_definition.0.request.0.conditional_formats.# = 2",
	"widget.8.query_value_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.8.query_value_definition.0.request.0.conditional_formats.0.value = 2",
	"widget.8.query_value_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.8.query_value_definition.0.request.0.conditional_formats.0.metric = system.load.1",
	"widget.8.query_value_definition.0.request.0.conditional_formats.1.comparator = >",
	"widget.8.query_value_definition.0.request.0.conditional_formats.1.value = 2.2",
	"widget.8.query_value_definition.0.request.0.conditional_formats.1.palette = white_on_red",
	"widget.8.query_value_definition.0.request.0.conditional_formats.1.metric = system.load.1",
	"widget.8.query_value_definition.0.autoscale = true",
	"widget.8.query_value_definition.0.custom_unit = xx",
	"widget.8.query_value_definition.0.precision = 4",
	"widget.8.query_value_definition.0.title = Widget Title",
	"widget.8.query_value_definition.0.time.live_span = 1h",
	// Scatterplot widget
	"widget.9.scatterplot_definition.0.request.0.x.0.q = avg:system.cpu.user{*} by {service, account}",
	"widget.9.scatterplot_definition.0.request.0.x.0.aggregator = max",
	"widget.9.scatterplot_definition.0.request.0.y.0.q = avg:system.mem.used{*} by {service, account}",
	"widget.9.scatterplot_definition.0.request.0.y.0.aggregator = min",
	"widget.9.scatterplot_definition.0.color_by_groups.# = 2",
	"widget.9.scatterplot_definition.0.color_by_groups.0 = account",
	"widget.9.scatterplot_definition.0.color_by_groups.1 = apm-role-group",
	"widget.9.scatterplot_definition.0.xaxis.0.include_zero = true",
	"widget.9.scatterplot_definition.0.xaxis.0.label = x",
	"widget.9.scatterplot_definition.0.xaxis.0.max = 2000",
	"widget.9.scatterplot_definition.0.xaxis.0.min = 1",
	"widget.9.scatterplot_definition.0.xaxis.0.scale = pow",
	"widget.9.scatterplot_definition.0.yaxis.0.include_zero = false",
	"widget.9.scatterplot_definition.0.yaxis.0.label = y",
	"widget.9.scatterplot_definition.0.yaxis.0.max = 2222",
	"widget.9.scatterplot_definition.0.yaxis.0.min = 5",
	"widget.9.scatterplot_definition.0.yaxis.0.scale = log",
	"widget.9.scatterplot_definition.0.title = Widget Title",
	"widget.9.scatterplot_definition.0.time.live_span = 1h",
	// Timeseries widget
	"widget.10.timeseries_definition.0.request.0.q = avg:system.cpu.user{app:general} by {env}",
	"widget.10.timeseries_definition.0.request.0.display_type = line",
	"widget.10.timeseries_definition.0.request.0.style.0.palette = warm",
	"widget.10.timeseries_definition.0.request.0.style.0.line_type = dashed",
	"widget.10.timeseries_definition.0.request.0.style.0.line_width = thin",
	"widget.10.timeseries_definition.0.request.0.metadata.0.expression = avg:system.cpu.user{app:general} by {env}",
	"widget.10.timeseries_definition.0.request.0.metadata.0.alias_name = Alpha",
	"widget.10.timeseries_definition.0.request.1.log_query.0.index = mcnulty",
	"widget.10.timeseries_definition.0.request.1.log_query.0.compute.aggregation = count",
	"widget.10.timeseries_definition.0.request.1.log_query.0.compute.facet = @duration",
	"widget.10.timeseries_definition.0.request.1.log_query.0.compute.interval = 5000",
	"widget.10.timeseries_definition.0.request.1.log_query.0.search.query = status:info",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.# = 1",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.facet = host",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.limit = 10",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.sort.aggregation = avg",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.sort.facet = @duration",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.sort.order = desc",
	"widget.10.timeseries_definition.0.request.1.display_type = area",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.index = apm-search",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.compute.aggregation = count",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.compute.facet = @duration",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.compute.interval = 5000",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.search.query = type:web",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.# = 1",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.facet = resource_name",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.limit = 50",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.aggregation = avg",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.facet = @string_query.interval",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.order = desc",
	"widget.10.timeseries_definition.0.request.2.display_type = bars",
	"widget.10.timeseries_definition.0.request.3.process_query.0.metric = process.stat.cpu.total_pct",
	"widget.10.timeseries_definition.0.request.3.process_query.0.search_by = error",
	"widget.10.timeseries_definition.0.request.3.process_query.0.filter_by.# = 1",
	"widget.10.timeseries_definition.0.request.3.process_query.0.filter_by.0 = active",
	"widget.10.timeseries_definition.0.request.3.process_query.0.limit = 50",
	"widget.10.timeseries_definition.0.request.3.display_type = area",
	"widget.10.timeseries_definition.0.request.4.security_query.0.index = signal",
	"widget.10.timeseries_definition.0.request.4.security_query.0.compute.aggregation = count",
	"widget.10.timeseries_definition.0.request.4.security_query.0.search.query = status:(high OR critical)",
	"widget.10.timeseries_definition.0.request.4.security_query.0.group_by.0.facet = status",
	"widget.10.timeseries_definition.0.request.4.display_type = bars",
	"widget.10.timeseries_definition.0.request.5.rum_query.0.index = rum",
	"widget.10.timeseries_definition.0.request.5.rum_query.0.compute.aggregation = count",
	"widget.10.timeseries_definition.0.request.5.rum_query.0.search.query = status:info",
	"widget.10.timeseries_definition.0.request.5.rum_query.0.group_by.0.facet = service",
	"widget.10.timeseries_definition.0.request.5.display_type = bars",
	"widget.10.timeseries_definition.0.marker.# = 2",
	"widget.10.timeseries_definition.0.marker.0.display_type = error dashed",
	"widget.10.timeseries_definition.0.marker.0.label =  z=6 ",
	"widget.10.timeseries_definition.0.marker.0.value = y=4",
	"widget.10.timeseries_definition.0.marker.1.display_type = ok solid",
	"widget.10.timeseries_definition.0.marker.1.label =  x=8 ",
	"widget.10.timeseries_definition.0.marker.1.value = 10 < y < 999",
	"widget.10.timeseries_definition.0.title = Widget Title",
	"widget.10.timeseries_definition.0.show_legend = true",
	"widget.10.timeseries_definition.0.legend_size = 2",
	"widget.10.timeseries_definition.0.time.live_span = 1h",
	"widget.10.timeseries_definition.0.event.0.q = sources:test tags:1",
	"widget.10.timeseries_definition.0.event.1.q = sources:test tags:2",
	"widget.10.timeseries_definition.0.yaxis.0.scale = log",
	"widget.10.timeseries_definition.0.yaxis.0.include_zero = false",
	"widget.10.timeseries_definition.0.yaxis.0.max = 100",
	// Toplist widget
	"widget.11.toplist_definition.0.request.0.q = avg:system.cpu.user{app:general} by {env}",
	"widget.11.toplist_definition.0.request.0.conditional_formats.# = 2",
	"widget.11.toplist_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.11.toplist_definition.0.request.0.conditional_formats.0.value = 2",
	"widget.11.toplist_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.11.toplist_definition.0.request.0.conditional_formats.1.comparator = >",
	"widget.11.toplist_definition.0.request.0.conditional_formats.1.value = 2.2",
	"widget.11.toplist_definition.0.request.0.conditional_formats.1.palette = white_on_red",
	"widget.11.toplist_definition.0.title = Widget Title",
	// Group widget
	"widget.12.group_definition.0.layout_type = ordered",
	"widget.12.group_definition.0.title = Group Widget",
	"widget.12.group_definition.0.widget.# = 2",
	// Inner Note widget
	"widget.12.group_definition.0.widget.0.note_definition.0.content = cluster note widget",
	"widget.12.group_definition.0.widget.0.note_definition.0.background_color = yellow",
	"widget.12.group_definition.0.widget.0.note_definition.0.font_size = 16",
	"widget.12.group_definition.0.widget.0.note_definition.0.text_align = left",
	"widget.12.group_definition.0.widget.0.note_definition.0.show_tick = false",
	"widget.12.group_definition.0.widget.0.note_definition.0.tick_edge = left",
	"widget.12.group_definition.0.widget.0.note_definition.0.tick_pos = 50%",
	// Inner Alert Graph widget
	"widget.12.group_definition.0.widget.1.alert_graph_definition.0.alert_id = 123",
	"widget.12.group_definition.0.widget.1.alert_graph_definition.0.viz_type = toplist",
	"widget.12.group_definition.0.widget.1.alert_graph_definition.0.title = Alert Graph",
	"widget.12.group_definition.0.widget.1.alert_graph_definition.0.time.live_span = 1h",
	// Service Level Objective widget
	"widget.13.service_level_objective_definition.0.title = Widget Title",
	"widget.13.service_level_objective_definition.0.view_type = detail",
	"widget.13.service_level_objective_definition.0.slo_id = 56789",
	"widget.13.service_level_objective_definition.0.show_error_budget = true",
	"widget.13.service_level_objective_definition.0.view_mode = overall",
	"widget.13.service_level_objective_definition.0.time_windows.# = 2",
	"widget.13.service_level_objective_definition.0.time_windows.0 = 7d",
	"widget.13.service_level_objective_definition.0.time_windows.1 = previous_week",
	// Query Table widget
	"widget.14.query_table_definition.0.request.0.q = avg:system.load.1{env:staging} by {account}",
	"widget.14.query_table_definition.0.request.0.conditional_formats.# = 2",
	"widget.14.query_table_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.14.query_table_definition.0.request.0.conditional_formats.0.value = 2",
	"widget.14.query_table_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.14.query_table_definition.0.request.0.conditional_formats.1.comparator = >",
	"widget.14.query_table_definition.0.request.0.conditional_formats.1.value = 2.2",
	"widget.14.query_table_definition.0.request.0.conditional_formats.1.palette = white_on_red",
	"widget.14.query_table_definition.0.request.0.aggregator = sum",
	"widget.14.query_table_definition.0.request.0.limit = 10",
	"widget.14.query_table_definition.0.title = Widget Title",
	"widget.14.query_table_definition.0.time.live_span = 1h",
	// Template Variables
	"template_variable.# = 2",
	"template_variable.0.name = var_1",
	"template_variable.0.prefix = host",
	"template_variable.0.default = aws",
	"template_variable.1.name = var_2",
	"template_variable.1.prefix = service_name",
	"template_variable.1.default = autoscaling",
	"description = Created using the Datadog provider in Terraform",

	// Template Variable Presets
	"template_variable_preset.# = 2",
	"template_variable_preset.0.name = preset_1",
	"template_variable_preset.0.template_variable.0.name = var_1",
	"template_variable_preset.0.template_variable.0.value = var_1_value",
	"template_variable_preset.0.template_variable.1.name = var_2",
	"template_variable_preset.0.template_variable.1.value = var_2_value",
	"template_variable_preset.1.name = preset_2",
	"template_variable_preset.1.template_variable.0.name = var_1",
	"template_variable_preset.1.template_variable.0.value = var_1_value",
}

var datadogFreeDashboardAsserts = []string{
	// Dashboard metadata
	"description = Created using the Datadog provider in Terraform",
	"layout_type = free",
	"is_read_only = false",
	"widget.# = 8",

	// Event Stream widget
	"widget.0.event_stream_definition.0.query = *",
	"widget.0.event_stream_definition.0.event_size = l",
	"widget.0.event_stream_definition.0.title = Widget Title",
	"widget.0.event_stream_definition.0.title_size = 16",
	"widget.0.event_stream_definition.0.title_align = left",
	"widget.0.event_stream_definition.0.time.live_span = 1h",
	"widget.0.layout.height = 43",
	"widget.0.layout.width = 32",
	"widget.0.layout.x = 5",
	"widget.0.layout.y = 5",
	// Event Timeline widget
	"widget.1.event_timeline_definition.0.query = *",
	"widget.1.event_timeline_definition.0.title = Widget Title",
	"widget.1.event_timeline_definition.0.title_align = left",
	"widget.1.event_timeline_definition.0.title_size = 16",
	"widget.1.event_timeline_definition.0.time.live_span = 1h",
	"widget.1.layout.height = 9",
	"widget.1.layout.width = 65",
	"widget.1.layout.x = 42",
	"widget.1.layout.y = 73",
	// Free Text widget
	"widget.2.free_text_definition.0.text = free text content",
	"widget.2.free_text_definition.0.color = #d00",
	"widget.2.free_text_definition.0.font_size = 88",
	"widget.2.free_text_definition.0.text_align = left",
	"widget.2.layout.height = 20",
	"widget.2.layout.width = 30",
	"widget.2.layout.x = 42",
	"widget.2.layout.y = 5",
	// Iframe widget
	"widget.3.iframe_definition.0.url = http://google.com",
	// Image widget
	"widget.4.image_definition.0.url = https://images.pexels.com/photos/67636/rose-blue-flower-rose-blooms-67636.jpeg?auto=compress&cs=tinysrgb&h=350",
	"widget.4.image_definition.0.sizing = fit",
	"widget.4.image_definition.0.margin = small",
	"widget.4.layout.height = 20",
	"widget.4.layout.width = 30",
	"widget.4.layout.x = 77",
	"widget.4.layout.y = 7",
	// Log Stream widget
	"widget.5.log_stream_definition.0.logset = 19",
	"widget.5.log_stream_definition.0.query = error",
	"widget.5.log_stream_definition.0.columns.# = 3",
	"widget.5.log_stream_definition.0.columns.0 = core_host",
	"widget.5.log_stream_definition.0.columns.1 = core_service",
	"widget.5.log_stream_definition.0.columns.2 = tag_source",
	"widget.5.log_stream_definition.0.show_date_column = true",
	"widget.5.log_stream_definition.0.show_message_column = true",
	"widget.5.log_stream_definition.0.message_display = expanded-md",
	"widget.5.log_stream_definition.0.sort.0.column = time",
	"widget.5.log_stream_definition.0.sort.0.order = desc",
	"widget.5.layout.height = 36",
	"widget.5.layout.width = 32",
	"widget.5.layout.x = 5",
	"widget.5.layout.y = 51",
	// Manage Status widget
	"widget.6.manage_status_definition.0.color_preference = text",
	"widget.6.manage_status_definition.0.count = 50",
	"widget.6.manage_status_definition.0.display_format = countsAndList",
	"widget.6.manage_status_definition.0.hide_zero_counts = true",
	"widget.6.manage_status_definition.0.query = type:metric",
	"widget.6.manage_status_definition.0.show_last_triggered = true",
	"widget.6.manage_status_definition.0.sort = status,asc",
	"widget.6.manage_status_definition.0.start = 0",
	"widget.6.manage_status_definition.0.summary_type = monitors",
	"widget.6.manage_status_definition.0.title = Widget Title",
	"widget.6.manage_status_definition.0.title_align = left",
	"widget.6.manage_status_definition.0.title_size = 16",
	"widget.6.layout.height = 40",
	"widget.6.layout.width = 30",
	"widget.6.layout.x = 112",
	"widget.6.layout.y = 55",
	// Trace Service widget
	"widget.7.trace_service_definition.0.display_format = three_column",
	"widget.7.trace_service_definition.0.env = datad0g.com",
	"widget.7.trace_service_definition.0.service = alerting-cassandra",
	"widget.7.trace_service_definition.0.show_breakdown = true",
	"widget.7.trace_service_definition.0.show_distribution = true",
	"widget.7.trace_service_definition.0.show_errors = true",
	"widget.7.trace_service_definition.0.show_hits = true",
	"widget.7.trace_service_definition.0.show_latency = false",
	"widget.7.trace_service_definition.0.show_resource_list = false",
	"widget.7.trace_service_definition.0.size_format = large",
	"widget.7.trace_service_definition.0.span_name = cassandra.query",
	"widget.7.trace_service_definition.0.title = alerting-cassandra #env:datad0g.com",
	"widget.7.trace_service_definition.0.title_align = center",
	"widget.7.trace_service_definition.0.title_size = 13",
	"widget.7.trace_service_definition.0.time.live_span = 1h",
	// Template Variables
	"template_variable.# = 2",
	"template_variable.0.default = aws",
	"template_variable.0.name = var_1",
	"template_variable.0.prefix = host",
	"template_variable.1.default = autoscaling",
	"template_variable.1.name = var_2",
	"template_variable.1.prefix = service_name",

	// Template Variable Presets
	"template_variable_preset.# = 2",
	"template_variable_preset.0.name = preset_1",
	"template_variable_preset.0.template_variable.0.name = var_1",
	"template_variable_preset.0.template_variable.0.value = var_1_value",
	"template_variable_preset.0.template_variable.1.name = var_2",
	"template_variable_preset.0.template_variable.1.value = var_2_value",
	"template_variable_preset.1.name = preset_2",
	"template_variable_preset.1.template_variable.0.name = var_1",
	"template_variable_preset.1.template_variable.0.value = var_1_value",
}

func TestAccDatadogDashboard_update(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	dbName := uniqueEntityName(clock, t)
	asserts := datadogOrderedDashboardAsserts
	asserts = append(asserts, fmt.Sprintf("title = %s", dbName))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogOrderedDashboardConfig(dbName),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.ordered_dashboard", checkDashboardExists(accProvider), asserts)...,
				),
			},
		},
	})
}

func TestAccDatadogFreeDashboard(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	dbName := uniqueEntityName(clock, t)
	asserts := datadogFreeDashboardAsserts
	asserts = append(asserts, fmt.Sprintf("title = %s", dbName))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogFreeDashboardConfig(dbName),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.free_dashboard", checkDashboardExists(accProvider), asserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboard_import(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	dbName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogOrderedDashboardConfig(dbName),
			},
			{
				ResourceName:      "datadog_dashboard.ordered_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: datadogFreeDashboardConfig(dbName),
			},
			{
				ResourceName:      "datadog_dashboard.free_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkDashboardExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, _, err := datadogClientV1.DashboardsApi.GetDashboard(authV1, r.Primary.ID).Execute(); err != nil {
				return fmt.Errorf("received an error retrieving dashboard1 %s", err)
			}
		}
		return nil
	}
}

func checkDashboardDestroy(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, _, err := datadogClientV1.DashboardsApi.GetDashboard(authV1, r.Primary.ID).Execute(); err != nil {
				if strings.Contains(err.Error(), "404 Not Found") {
					continue
				}
				return fmt.Errorf("received an error retrieving dashboard2 %s", err)
			}
			return fmt.Errorf("dashboard still exists")
		}
		return nil
	}
}

func testAccDatadogDashboardWidgetUtil(t *testing.T, config string, name string, assertions []string) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	uniq := uniqueEntityName(clock, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)
	config = replacer.Replace(config)
	for i := range assertions {
		assertions[i] = replacer.Replace(assertions[i])
	}
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs(name, checkDashboardExists(accProvider), assertions)...,
				),
			},
		},
	})
}

func testAccDatadogDashboardWidgetUtil_import(t *testing.T, config string, name string) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	uniq := uniqueEntityName(clock, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)
	config = replacer.Replace(config)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
