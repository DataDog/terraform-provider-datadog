package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
			live_span = "1h"
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
			live_span = "1h"
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
			live_span = "1h"
		}
	}
	widget {
		check_status_definition {
			check = "aws.ecs.agent_connected"
			grouping = "cluster"
			group_by = ["account", "cluster"]
			tags = ["account:demo", "cluster:awseb-ruthebdog-env-8-dn3m6u3gvk"]
			title = "Widget Title"
			live_span = "1h"
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
			live_span = "1h"
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
			live_span = "1h"
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
			live_span = "1h"
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
					compute_query {
						aggregation = "count"
						facet = "@duration"
						interval = 5000
					}
					search_query = "status:info"
					group_by {
						facet = "host"
						limit = 10
						sort_query {
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
					compute_query {
						aggregation = "count"
						facet = "@duration"
						interval = 5000
					}
					search_query = "type:web"
					group_by {
						facet = "resource_name"
						limit = 50
						sort_query {
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
					compute_query {
						aggregation = "count"
					}
					search_query = "status:(high OR critical)"
					group_by {
						facet = "status"
					}
				}
				display_type = "bars"
			}
			request {
				rum_query {
					index = "rum"
					compute_query {
						aggregation = "count"
					}
					search_query = "status:info"
					group_by {
						facet = "service"
					}
				}
				display_type = "bars"
			}
			request {
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
			live_span = "1h"
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
					live_span = "1h"
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
			live_span = "1h"
		}
	}
	widget {
		query_table_definition {
			request {
				apm_stats_query {
					service = "foo"
					name = "bar"
					env = "staging"
					primary_tag = "datacenter:*"
					row_type = "resource"
					columns {
						name = "Hits"
					}
				}
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
	template_variable_preset {
		name = "preset_3"
}
}`, uniq)
}

func datadogSimpleOrderedDashboardConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "simple_dashboard" {
	title         = "%s"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = true
	widget {
		alert_graph_definition {
			alert_id = "895605"
			viz_type = "timeseries"
			title = "Widget Title"
			live_span = "1h"
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
	template_variable_preset {
		name = "preset_3" 
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
			live_span = "1h"
		}
		widget_layout {
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
			live_span = "1h"
		}
		widget_layout {
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
		widget_layout {
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
		widget_layout {
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
		widget_layout {
			height = 20
			width = 30
			x = 77
			y = 7
		}
	}
	widget {
		log_stream_definition {
			indexes = ["main"]
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
		widget_layout {
			height = 36
			width = 32
			x = 5
			y = 51
		}
	}
	widget {
		manage_status_definition {
			color_preference = "text"
			display_format = "countsAndList"
			hide_zero_counts = true
			query = "type:metric"
			show_last_triggered = true
			sort = "status,asc"
			summary_type = "monitors"
			title = "Widget Title"
			title_size = 16
			title_align = "left"
		}
		widget_layout {
			height = 40
			width = 30
			x = 112
			y = 55
		}
	}
	widget {
		trace_service_definition {
			display_format = "three_column"
			env = "datadog.com"
			service = "alerting-cassandra"
			show_breakdown = true
			show_distribution = true
			show_errors = true
			show_hits = true
			show_latency = false
			show_resource_list = false
			size_format = "large"
			span_name = "cassandra.query"
			title = "alerting-cassandra #env:datadog.com"
			title_align = "center"
			title_size = "13"
			live_span = "1h"
		}
		widget_layout {
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
	template_variable_preset {
		name = "preset_3"
	}
}`, uniq)
}

func datadogSimpleFreeDashboardConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "simple_dashboard" {
	title         = "%s"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = true
	widget {
		alert_graph_definition {
			alert_id = "895605"
			viz_type = "timeseries"
			title = "Widget Title"
			live_span = "1h"
		}
		widget_layout {
			height = 43
			width = 32
			x = 5
			y = 5
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
	template_variable_preset {
		name = "preset_3"
	}
}`, uniq)
}

var datadogSimpleOrderedDashboardAsserts = []string{
	// Dashboard metadata
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"is_read_only = true",
	"widget.# = 1",
	// Alert Graph widget
	"widget.0.alert_graph_definition.0.alert_id = 895605",
	"widget.0.alert_graph_definition.0.viz_type = timeseries",
	"widget.0.alert_graph_definition.0.title = Widget Title",
	"widget.0.alert_graph_definition.0.live_span = 1h",
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
	"template_variable_preset.# = 3",
	"template_variable_preset.0.name = preset_1",
	"template_variable_preset.0.template_variable.0.name = var_1",
	"template_variable_preset.0.template_variable.0.value = var_1_value",
	"template_variable_preset.0.template_variable.1.name = var_2",
	"template_variable_preset.0.template_variable.1.value = var_2_value",
	"template_variable_preset.1.name = preset_2",
	"template_variable_preset.1.template_variable.0.name = var_1",
	"template_variable_preset.1.template_variable.0.value = var_1_value",
	"template_variable_preset.2.name = preset_3",
}

var datadogSimpleFreeDashboardAsserts = []string{
	// Dashboard metadata
	"description = Created using the Datadog provider in Terraform",
	"layout_type = free",
	"is_read_only = true",
	"widget.# = 1",
	// Alert Graph widget
	"widget.0.alert_graph_definition.0.alert_id = 895605",
	"widget.0.alert_graph_definition.0.viz_type = timeseries",
	"widget.0.alert_graph_definition.0.title = Widget Title",
	"widget.0.alert_graph_definition.0.live_span = 1h",
	"widget.0.widget_layout.0.height = 43",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.widget_layout.0.y = 5",
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
	"template_variable_preset.# = 3",
	"template_variable_preset.0.name = preset_1",
	"template_variable_preset.0.template_variable.0.name = var_1",
	"template_variable_preset.0.template_variable.0.value = var_1_value",
	"template_variable_preset.0.template_variable.1.name = var_2",
	"template_variable_preset.0.template_variable.1.value = var_2_value",
	"template_variable_preset.1.name = preset_2",
	"template_variable_preset.1.template_variable.0.name = var_1",
	"template_variable_preset.1.template_variable.0.value = var_1_value",
	"template_variable_preset.2.name = preset_3",
}

var datadogOrderedDashboardAsserts = []string{
	// Dashboard metadata
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"is_read_only = true",
	"widget.# = 16",
	// Alert Graph widget
	"widget.0.alert_graph_definition.0.alert_id = 895605",
	"widget.0.alert_graph_definition.0.viz_type = timeseries",
	"widget.0.alert_graph_definition.0.title = Widget Title",
	"widget.0.alert_graph_definition.0.live_span = 1h",
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
	"widget.2.change_definition.0.live_span = 1h",
	// Distribution widget
	"widget.3.distribution_definition.0.request.0.q = avg:system.load.1{env:staging} by {account}",
	"widget.3.distribution_definition.0.request.0.style.0.palette = warm",
	"widget.3.distribution_definition.0.title = Widget Title",
	"widget.3.distribution_definition.0.live_span = 1h",
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
	"widget.4.check_status_definition.0.live_span = 1h",
	// Heatmap widget
	"widget.5.heatmap_definition.0.request.0.q = avg:system.load.1{env:staging} by {account}",
	"widget.5.heatmap_definition.0.request.0.style.0.palette = warm",
	"widget.5.heatmap_definition.0.yaxis.0.min = 1",
	"widget.5.heatmap_definition.0.yaxis.0.max = 2",
	"widget.5.heatmap_definition.0.yaxis.0.include_zero = true",
	"widget.5.heatmap_definition.0.yaxis.0.scale = sqrt",
	"widget.5.heatmap_definition.0.title = Widget Title",
	"widget.5.heatmap_definition.0.live_span = 1h",
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
	"widget.8.query_value_definition.0.live_span = 1h",
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
	"widget.9.scatterplot_definition.0.live_span = 1h",
	// Timeseries widget
	"widget.10.timeseries_definition.0.request.0.q = avg:system.cpu.user{app:general} by {env}",
	"widget.10.timeseries_definition.0.request.0.display_type = line",
	"widget.10.timeseries_definition.0.request.0.style.0.palette = warm",
	"widget.10.timeseries_definition.0.request.0.style.0.line_type = dashed",
	"widget.10.timeseries_definition.0.request.0.style.0.line_width = thin",
	"widget.10.timeseries_definition.0.request.0.metadata.0.expression = avg:system.cpu.user{app:general} by {env}",
	"widget.10.timeseries_definition.0.request.0.metadata.0.alias_name = Alpha",
	"widget.10.timeseries_definition.0.request.1.log_query.0.index = mcnulty",
	"widget.10.timeseries_definition.0.request.1.log_query.0.compute_query.0.aggregation = count",
	"widget.10.timeseries_definition.0.request.1.log_query.0.compute_query.0.facet = @duration",
	"widget.10.timeseries_definition.0.request.1.log_query.0.compute_query.0.interval = 5000",
	"widget.10.timeseries_definition.0.request.1.log_query.0.search_query = status:info",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.# = 1",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.facet = host",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.limit = 10",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.sort_query.0.aggregation = avg",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.sort_query.0.facet = @duration",
	"widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.sort_query.0.order = desc",
	"widget.10.timeseries_definition.0.request.1.display_type = area",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.index = apm-search",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.compute_query.0.aggregation = count",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.compute_query.0.facet = @duration",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.compute_query.0.interval = 5000",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.search_query = type:web",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.# = 1",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.facet = resource_name",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.limit = 50",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort_query.0.aggregation = avg",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort_query.0.facet = @string_query.interval",
	"widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort_query.0.order = desc",
	"widget.10.timeseries_definition.0.request.2.display_type = bars",
	"widget.10.timeseries_definition.0.request.3.process_query.0.metric = process.stat.cpu.total_pct",
	"widget.10.timeseries_definition.0.request.3.process_query.0.search_by = error",
	"widget.10.timeseries_definition.0.request.3.process_query.0.filter_by.# = 1",
	"widget.10.timeseries_definition.0.request.3.process_query.0.filter_by.0 = active",
	"widget.10.timeseries_definition.0.request.3.process_query.0.limit = 50",
	"widget.10.timeseries_definition.0.request.3.display_type = area",
	"widget.10.timeseries_definition.0.request.4.security_query.0.index = signal",
	"widget.10.timeseries_definition.0.request.4.security_query.0.compute_query.0.aggregation = count",
	"widget.10.timeseries_definition.0.request.4.security_query.0.search_query = status:(high OR critical)",
	"widget.10.timeseries_definition.0.request.4.security_query.0.group_by.0.facet = status",
	"widget.10.timeseries_definition.0.request.4.display_type = bars",
	"widget.10.timeseries_definition.0.request.5.rum_query.0.index = rum",
	"widget.10.timeseries_definition.0.request.5.rum_query.0.compute_query.0.aggregation = count",
	"widget.10.timeseries_definition.0.request.5.rum_query.0.search_query = status:info",
	"widget.10.timeseries_definition.0.request.5.rum_query.0.group_by.0.facet = service",
	"widget.10.timeseries_definition.0.request.5.display_type = bars",
	"widget.10.timeseries_definition.0.request.6.audit_query.0.index = *",
	"widget.10.timeseries_definition.0.request.6.audit_query.0.compute_query.0.aggregation = count",
	"widget.10.timeseries_definition.0.request.6.audit_query.0.search_query =",
	"widget.10.timeseries_definition.0.request.6.audit_query.0.group_by.0.facet = @metadata.api_key.id",
	"widget.10.timeseries_definition.0.request.6.audit_query.0.group_by.0.sort_query.0.aggregation = count",
	"widget.10.timeseries_definition.0.request.6.audit_query.0.group_by.0.sort_query.0.order = desc",
	"widget.10.timeseries_definition.0.request.6.audit_query.0.group_by.0.limit = 10",
	"widget.10.timeseries_definition.0.request.6.display_type = line",
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
	"widget.10.timeseries_definition.0.live_span = 1h",
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
	"widget.12.group_definition.0.widget.1.alert_graph_definition.0.live_span = 1h",
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
	"widget.14.query_table_definition.0.live_span = 1h",
	"widget.15.query_table_definition.0.request.0.apm_stats_query.0.service = foo",
	"widget.15.query_table_definition.0.request.0.apm_stats_query.0.name = bar",
	"widget.15.query_table_definition.0.request.0.apm_stats_query.0.env = staging",
	"widget.15.query_table_definition.0.request.0.apm_stats_query.0.primary_tag = datacenter:*",
	"widget.15.query_table_definition.0.request.0.apm_stats_query.0.row_type = resource",
	"widget.15.query_table_definition.0.request.0.apm_stats_query.0.columns.# = 1",
	"widget.15.query_table_definition.0.request.0.apm_stats_query.0.columns.0.name = Hits",
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
	"template_variable_preset.# = 3",
	"template_variable_preset.0.name = preset_1",
	"template_variable_preset.0.template_variable.0.name = var_1",
	"template_variable_preset.0.template_variable.0.value = var_1_value",
	"template_variable_preset.0.template_variable.1.name = var_2",
	"template_variable_preset.0.template_variable.1.value = var_2_value",
	"template_variable_preset.1.name = preset_2",
	"template_variable_preset.1.template_variable.0.name = var_1",
	"template_variable_preset.1.template_variable.0.value = var_1_value",
	"template_variable_preset.2.name = preset_3",
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
	"widget.0.event_stream_definition.0.live_span = 1h",
	"widget.0.widget_layout.0.height = 43",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.widget_layout.0.y = 5",
	// Event Timeline widget
	"widget.1.event_timeline_definition.0.query = *",
	"widget.1.event_timeline_definition.0.title = Widget Title",
	"widget.1.event_timeline_definition.0.title_align = left",
	"widget.1.event_timeline_definition.0.title_size = 16",
	"widget.1.event_timeline_definition.0.live_span = 1h",
	"widget.1.widget_layout.0.height = 9",
	"widget.1.widget_layout.0.width = 65",
	"widget.1.widget_layout.0.x = 42",
	"widget.1.widget_layout.0.y = 73",
	// Free Text widget
	"widget.2.free_text_definition.0.text = free text content",
	"widget.2.free_text_definition.0.color = #d00",
	"widget.2.free_text_definition.0.font_size = 88",
	"widget.2.free_text_definition.0.text_align = left",
	"widget.2.widget_layout.0.height = 20",
	"widget.2.widget_layout.0.width = 30",
	"widget.2.widget_layout.0.x = 42",
	"widget.2.widget_layout.0.y = 5",
	// Iframe widget
	"widget.3.iframe_definition.0.url = http://google.com",
	// Image widget
	"widget.4.image_definition.0.url = https://images.pexels.com/photos/67636/rose-blue-flower-rose-blooms-67636.jpeg?auto=compress&cs=tinysrgb&h=350",
	"widget.4.image_definition.0.sizing = fit",
	"widget.4.image_definition.0.margin = small",
	"widget.4.widget_layout.0.height = 20",
	"widget.4.widget_layout.0.width = 30",
	"widget.4.widget_layout.0.x = 77",
	"widget.4.widget_layout.0.y = 7",
	// Log Stream widget
	"widget.5.log_stream_definition.0.indexes.0 = main",
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
	"widget.5.widget_layout.0.height = 36",
	"widget.5.widget_layout.0.width = 32",
	"widget.5.widget_layout.0.x = 5",
	"widget.5.widget_layout.0.y = 51",
	// Manage Status widget
	"widget.6.manage_status_definition.0.color_preference = text",
	"widget.6.manage_status_definition.0.display_format = countsAndList",
	"widget.6.manage_status_definition.0.hide_zero_counts = true",
	"widget.6.manage_status_definition.0.query = type:metric",
	"widget.6.manage_status_definition.0.show_last_triggered = true",
	"widget.6.manage_status_definition.0.sort = status,asc",
	"widget.6.manage_status_definition.0.summary_type = monitors",
	"widget.6.manage_status_definition.0.title = Widget Title",
	"widget.6.manage_status_definition.0.title_align = left",
	"widget.6.manage_status_definition.0.title_size = 16",
	"widget.6.manage_status_definition.0.show_priority = false",
	"widget.6.widget_layout.0.height = 40",
	"widget.6.widget_layout.0.width = 30",
	"widget.6.widget_layout.0.x = 112",
	"widget.6.widget_layout.0.y = 55",
	// Trace Service widget
	"widget.7.trace_service_definition.0.display_format = three_column",
	"widget.7.trace_service_definition.0.env = datadog.com",
	"widget.7.trace_service_definition.0.service = alerting-cassandra",
	"widget.7.trace_service_definition.0.show_breakdown = true",
	"widget.7.trace_service_definition.0.show_distribution = true",
	"widget.7.trace_service_definition.0.show_errors = true",
	"widget.7.trace_service_definition.0.show_hits = true",
	"widget.7.trace_service_definition.0.show_latency = false",
	"widget.7.trace_service_definition.0.show_resource_list = false",
	"widget.7.trace_service_definition.0.size_format = large",
	"widget.7.trace_service_definition.0.span_name = cassandra.query",
	"widget.7.trace_service_definition.0.title = alerting-cassandra #env:datadog.com",
	"widget.7.trace_service_definition.0.title_align = center",
	"widget.7.trace_service_definition.0.title_size = 13",
	"widget.7.trace_service_definition.0.live_span = 1h",
	// Template Variables
	"template_variable.# = 2",
	"template_variable.0.default = aws",
	"template_variable.0.name = var_1",
	"template_variable.0.prefix = host",
	"template_variable.1.default = autoscaling",
	"template_variable.1.name = var_2",
	"template_variable.1.prefix = service_name",

	// Template Variable Presets
	"template_variable_preset.# = 3",
	"template_variable_preset.0.name = preset_1",
	"template_variable_preset.0.template_variable.0.name = var_1",
	"template_variable_preset.0.template_variable.0.value = var_1_value",
	"template_variable_preset.0.template_variable.1.name = var_2",
	"template_variable_preset.0.template_variable.1.value = var_2_value",
	"template_variable_preset.1.name = preset_2",
	"template_variable_preset.1.template_variable.0.name = var_1",
	"template_variable_preset.1.template_variable.0.value = var_1_value",
	"template_variable_preset.2.name = preset_3",
}

func TestAccDatadogDashboard_update(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	dbName := uniqueEntityName(ctx, t)
	asserts := datadogOrderedDashboardAsserts
	asserts = append(asserts, fmt.Sprintf("title = %s", dbName))
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.ordered_dashboard", checkDashboardExists(accProvider), asserts)
	for i := 0; i < 16; i++ {
		checks = append(checks, resource.TestCheckResourceAttrSet(
			"datadog_dashboard.ordered_dashboard", fmt.Sprintf("widget.%d.id", i)))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogOrderedDashboardConfig(dbName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogFreeDashboard(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	dbName := uniqueEntityName(ctx, t)
	asserts := datadogFreeDashboardAsserts
	asserts = append(asserts, fmt.Sprintf("title = %s", dbName))
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.free_dashboard", checkDashboardExists(accProvider), asserts)
	for i := 0; i < 8; i++ {
		checks = append(checks, resource.TestCheckResourceAttrSet(
			"datadog_dashboard.free_dashboard", fmt.Sprintf("widget.%d.id", i)))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogFreeDashboardConfig(dbName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogDashboardLayoutForceNew(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	dbName := uniqueEntityName(ctx, t)
	freeAsserts := datadogSimpleFreeDashboardAsserts
	freeAsserts = append(freeAsserts, fmt.Sprintf("title = %s", dbName))
	orderedAsserts := datadogSimpleOrderedDashboardAsserts
	orderedAsserts = append(orderedAsserts, fmt.Sprintf("title = %s", dbName))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogSimpleFreeDashboardConfig(dbName),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.simple_dashboard", checkDashboardExists(accProvider), freeAsserts)...,
				),
			},
			{
				Config: datadogSimpleOrderedDashboardConfig(dbName),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.simple_dashboard", checkDashboardExists(accProvider), orderedAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboard_import(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	dbName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
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

func checkDashboardExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_dashboard" && r.Type != "datadog_dashboard_json" {
				continue
			}
			if _, _, err := apiInstances.GetDashboardsApiV1().GetDashboard(auth, r.Primary.ID); err != nil {
				return fmt.Errorf("received an error retrieving dashboard1 %s", err)
			}
		}
		return nil
	}
}

func checkDashboardDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		err := utils.Retry(2, 10, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_dashboard" && r.Type != "datadog_dashboard_json" {
					continue
				}
				if _, httpResp, err := apiInstances.GetDashboardsApiV1().GetDashboard(auth, r.Primary.ID); err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Dashboard %s", err)}
				}
				return &utils.RetryableError{Prob: "Dashboard still exists"}
			}
			return nil
		})
		return err
	}
}

func testAccDatadogDashboardWidgetUtil(t *testing.T, config string, name string, assertions []string) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)
	config = replacer.Replace(config)
	for i := range assertions {
		assertions[i] = replacer.Replace(assertions[i])
	}
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
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

func testAccDatadogDashboardWidgetUtilImport(t *testing.T, config string, name string) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)
	config = replacer.Replace(config)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
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

func datadogOpenDashboardConfig(uniqueDashboardName string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "rbac_dashboard" {
	title            = "%s"
	description      = "Created using the Datadog provider in Terraform"
	layout_type      = "ordered"
	widget {
		note_definition {
			content = "note text"
		}
	}
}`, uniqueDashboardName)
}

var datadogOpenDashboardAsserts = []string{
	"is_read_only = false",
	"restricted_roles.# = 0",
}

func datadogAdminDashboardConfig(uniqueDashboardName string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "rbac_dashboard" {
	title            = "%s"
	description      = "Created using the Datadog provider in Terraform"
	layout_type      = "ordered"
	widget {
		note_definition {
			content = "note text"
		}
	}
	is_read_only     = true
}`, uniqueDashboardName)
}

var datadogAdminDashboardAsserts = []string{
	"is_read_only = true",
}

func datadogRbacDashboardConfig(uniqueDashboardName string, uniqueRoleName string) string {
	return fmt.Sprintf(`
resource "datadog_role" "rbac_role" {
	name = "%s"
}

resource "datadog_dashboard" "rbac_dashboard" {
	title            = "%s"
	description      = "Created using the Datadog provider in Terraform"
	layout_type      = "ordered"
	widget {
		note_definition {
			content = "note text"
		}
	}
	restricted_roles = ["${datadog_role.rbac_role.id}"]
}`, uniqueRoleName, uniqueDashboardName)
}

var datadogRbacDashboardAsserts = []string{
	"restricted_roles.# = 1",
}

func TestAccDatadogDashboardRbac_createOpen(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	asserts := datadogOpenDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.rbac_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogOpenDashboardConfig(boardName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogDashboardRbac_createAdmin(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	asserts := datadogAdminDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.rbac_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogAdminDashboardConfig(boardName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogDashboardRbac_createRbac(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	roleName := uniqueEntityName(ctx, t)
	asserts := datadogRbacDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.rbac_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogRbacDashboardConfig(boardName, roleName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogDashboardRbac_updateToAdmin(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	asserts := datadogAdminDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.rbac_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogOpenDashboardConfig(boardName),
			},
			{
				Config: datadogAdminDashboardConfig(boardName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogDashboardRbac_updateToRbac(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	roleName := uniqueEntityName(ctx, t)
	asserts := datadogRbacDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.rbac_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogOpenDashboardConfig(boardName),
			},
			{
				Config: datadogRbacDashboardConfig(boardName, roleName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogDashboardRbac_updateToOpen(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	roleName := uniqueEntityName(ctx, t)
	asserts := datadogOpenDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.rbac_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogAdminDashboardConfig(boardName),
			},
			{
				Config: datadogOpenDashboardConfig(boardName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			{
				Config: datadogRbacDashboardConfig(boardName, roleName),
			},
			{
				Config: datadogOpenDashboardConfig(boardName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogDashboardRbac_adminToRbac(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	roleName := uniqueEntityName(ctx, t)
	asserts := datadogRbacDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.rbac_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogAdminDashboardConfig(boardName),
			},
			{
				Config: datadogRbacDashboardConfig(boardName, roleName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func datadogMultiSizeLayoutFixedDashboardConfig(uniqueDashboardName string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "msl_fixed_dashboard" {
	title            = "%s"
	description      = "Created using the Datadog provider in Terraform"
	layout_type      = "ordered"
	reflow_type      = "fixed"
	widget {
		note_definition {
			content = "note 1"
		}
		widget_layout {
			width  = 6
			height = 2
			x      = 3
			y      = 0
		}  
	}
	widget {
		note_definition {
			content = "note 2"
		}
		widget_layout {
			width  = 6
			height = 2
			x      = 3
			y      = 2
			is_column_break = true
		}
	}
}`, uniqueDashboardName)
}

var datadogMslFixedDashboardAsserts = []string{
	// Dashboard metadata
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"reflow_type = fixed",
	"widget.# = 2",
	// First note layout
	"widget.0.note_definition.0.content = note 1",
	"widget.0.widget_layout.0.width = 6",
	"widget.0.widget_layout.0.height = 2",
	"widget.0.widget_layout.0.x = 3",
	"widget.0.widget_layout.0.y = 0",
	"widget.0.widget_layout.0.is_column_break = false",
	// Second note layout
	"widget.1.note_definition.0.content = note 2",
	"widget.1.widget_layout.0.width = 6",
	"widget.1.widget_layout.0.height = 2",
	"widget.1.widget_layout.0.x = 3",
	"widget.1.widget_layout.0.y = 2",
	"widget.1.widget_layout.0.is_column_break = true",
}

func datadogMultiSizeLayoutAutoDashboardConfig(uniqueDashboardName string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "msl_auto_dashboard" {
	title            = "%s"
	description      = "Created using the Datadog provider in Terraform"
	layout_type      = "ordered"
	reflow_type      = "auto"
	widget {
		note_definition {
			content = "note 1"
		}
	}
	widget {
		note_definition {
			content = "note 2"
		}
	}
}`, uniqueDashboardName)
}

var datadogMslAutoDashboardAsserts = []string{
	// Dashboard metadata
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"reflow_type = auto",
	"widget.# = 2",
	// First note layout
	"widget.0.note_definition.0.content = note 1",
	"widget.0.widget_layout.# = 0",
	// Second note layout
	"widget.1.note_definition.0.content = note 2",
	"widget.1.widget_layout.# = 0",
}

func TestAccDatadogDashboardMultiSizeLayout_createFixed(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	asserts := datadogMslFixedDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.msl_fixed_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogMultiSizeLayoutFixedDashboardConfig(boardName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogDashboardMultiSizeLayout_createAuto(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	asserts := datadogMslAutoDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.msl_auto_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogMultiSizeLayoutAutoDashboardConfig(boardName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func datadogDashboardNotifyListConfig(uniqueDashboardName string) string {
	return fmt.Sprintf(`
resource "datadog_user" "one" {
  email     = "z-user@example.com"
  name      = "Test User"
}
resource "datadog_user" "two" {
  email     = "a-user@example.com"
  name      = "Test User"
}
resource "datadog_user" "three" {
  email     = "k-user@example.com"
  name      = "Test User"
}

resource "datadog_dashboard" "ordered_dashboard" {
  title        = "%s"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "ordered"
  is_read_only = true
  notify_list  = [datadog_user.one.email, datadog_user.two.email, datadog_user.three.email]
  
  depends_on = [
    datadog_user.one,
    datadog_user.two,
    datadog_user.three,
  ]
}`, uniqueDashboardName)
}

var notifyListDashboardAsserts = []string{
	"notify_list.# = 3",
	"notify_list.TypeSet = z-user@example.com",
	"notify_list.TypeSet = a-user@example.com",
	"notify_list.TypeSet = k-user@example.com",
}

func TestAccDatadogDashboardNotifyListDiff(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	boardName := uniqueEntityName(ctx, t)
	asserts := notifyListDashboardAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_dashboard.ordered_dashboard", checkDashboardExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardNotifyListConfig(boardName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}
