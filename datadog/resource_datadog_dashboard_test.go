package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

const datadogDashboardConfig = `
resource "datadog_dashboard" "ordered_dashboard" {
  title         = "Acceptance Test Ordered Dashboard"
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
			tick_pos = "50%"
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
			}
			conditional_formats {
				comparator = ">"
				value = "2.2"
				palette = "white_on_red"
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
				q= "avg:system.cpu.user{app:general} by {env}"
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
			marker {
				display_type = "error dashed"
				label = " z=6 "
				value = "y = 4"
			}
			marker {
				display_type = "ok solid"
				value = "10 < y < 999"
				label = " x=8 "
			}
			title = "Widget Title"
			show_legend = true
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
					tick_pos = "50%"
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
}

resource "datadog_dashboard" "free_dashboard" {
	title         = "Acceptance Test Free Dashboard"
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
			sort = "status,asc"
			start = 0
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
}
`

func TestAccDatadogDashboard_update(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists,
					// Ordered layout dashboard

					// Dashboard metadata
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "title", "Acceptance Test Ordered Dashboard"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "description", "Created using the Datadog provider in Terraform"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "layout_type", "ordered"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "is_read_only", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.#", "13"),
					// Alert Graph widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.0.alert_graph_definition.0.alert_id", "895605"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.0.alert_graph_definition.0.viz_type", "timeseries"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.0.alert_graph_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.0.alert_graph_definition.0.time.live_span", "1h"),
					// Alert Value widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.1.alert_value_definition.0.alert_id", "895605"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.1.alert_value_definition.0.precision", "3"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.1.alert_value_definition.0.unit", "b"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.1.alert_value_definition.0.text_align", "center"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.1.alert_value_definition.0.title", "Widget Title"),
					// Change widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.2.change_definition.0.request.0.q", "avg:system.load.1{env:staging} by {account}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.2.change_definition.0.request.0.change_type", "absolute"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.2.change_definition.0.request.0.compare_to", "week_before"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.2.change_definition.0.request.0.increase_good", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.2.change_definition.0.request.0.order_by", "name"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.2.change_definition.0.request.0.order_dir", "desc"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.2.change_definition.0.request.0.show_present", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.2.change_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.2.change_definition.0.time.live_span", "1h"),
					// Distribution widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.3.distribution_definition.0.request.0.q", "avg:system.load.1{env:staging} by {account}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.3.distribution_definition.0.request.0.style.0.palette", "warm"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.3.distribution_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.3.distribution_definition.0.time.live_span", "1h"),
					// Check Status widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.check", "aws.ecs.agent_connected"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.grouping", "cluster"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.group_by.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.group_by.0", "account"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.group_by.1", "cluster"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.tags.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.tags.0", "account:demo"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.tags.1", "cluster:awseb-ruthebdog-env-8-dn3m6u3gvk"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.4.check_status_definition.0.time.live_span", "1h"),
					// Heatmap widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.5.heatmap_definition.0.request.0.q", "avg:system.load.1{env:staging} by {account}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.5.heatmap_definition.0.request.0.style.0.palette", "warm"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.5.heatmap_definition.0.yaxis.0.min", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.5.heatmap_definition.0.yaxis.0.max", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.5.heatmap_definition.0.yaxis.0.include_zero", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.5.heatmap_definition.0.yaxis.0.scale", "sqrt"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.5.heatmap_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.5.heatmap_definition.0.time.live_span", "1h"),
					// Hostmap widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.request.0.fill.0.q", "avg:system.load.1{*} by {host}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.request.0.size.0.q", "avg:memcache.uptime{*} by {host}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.node_type", "container"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.group.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.group.0", "host"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.group.1", "region"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.scope.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.scope.0", "region:us-east-1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.scope.1", "aws_account:727006795293"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.style.0.palette", "yellow_to_green"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.style.0.palette_flip", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.style.0.fill_min", "10"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.style.0.fill_max", "20"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.6.hostmap_definition.0.title", "Widget Title"),
					// Note widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.7.note_definition.0.content", "note text"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.7.note_definition.0.background_color", "pink"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.7.note_definition.0.font_size", "14"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.7.note_definition.0.text_align", "center"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.7.note_definition.0.show_tick", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.7.note_definition.0.tick_edge", "left"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.7.note_definition.0.tick_pos", "50%"),
					// Query valye widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.request.0.q", "avg:system.load.1{env:staging} by {account}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.request.0.aggregator", "sum"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.request.0.conditional_formats.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.request.0.conditional_formats.0.comparator", "<"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.request.0.conditional_formats.0.value", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.request.0.conditional_formats.0.palette", "white_on_green"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.request.0.conditional_formats.1.comparator", ">"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.request.0.conditional_formats.1.value", "2.2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.request.0.conditional_formats.1.palette", "white_on_red"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.autoscale", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.custom_unit", "xx"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.precision", "4"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.8.query_value_definition.0.time.live_span", "1h"),
					// Scatterplot widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.request.0.x.0.q", "avg:system.cpu.user{*} by {service, account}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.request.0.x.0.aggregator", "max"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.request.0.y.0.q", "avg:system.mem.used{*} by {service, account}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.request.0.y.0.aggregator", "min"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.color_by_groups.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.color_by_groups.0", "account"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.color_by_groups.1", "apm-role-group"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.xaxis.0.include_zero", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.xaxis.0.label", "x"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.xaxis.0.max", "2000"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.xaxis.0.min", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.xaxis.0.scale", "pow"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.yaxis.0.include_zero", "false"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.yaxis.0.label", "y"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.yaxis.0.max", "2222"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.yaxis.0.min", "5"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.yaxis.0.scale", "log"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.9.scatterplot_definition.0.time.live_span", "1h"),
					// Timeseries widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.0.q", "avg:system.cpu.user{app:general} by {env}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.0.display_type", "line"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.0.style.0.palette", "warm"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.0.style.0.line_type", "dashed"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.0.style.0.line_width", "thin"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.0.metadata.0.expression", "avg:system.cpu.user{app:general} by {env}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.0.metadata.0.alias_name", "Alpha"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.index", "mcnulty"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.compute.aggregation", "count"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.compute.facet", "@duration"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.compute.interval", "5000"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.search.query", "status:info"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.group_by.#", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.facet", "host"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.limit", "10"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.sort.aggregation", "avg"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.sort.facet", "@duration"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.log_query.0.group_by.0.sort.order", "desc"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.1.display_type", "area"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.index", "apm-search"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.compute.aggregation", "count"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.compute.facet", "@duration"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.compute.interval", "5000"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.search.query", "type:web"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.#", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.facet", "resource_name"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.limit", "50"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.aggregation", "avg"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.facet", "@string_query.interval"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.apm_query.0.group_by.0.sort.order", "desc"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.2.display_type", "bars"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.3.process_query.0.metric", "process.stat.cpu.total_pct"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.3.process_query.0.search_by", "error"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.3.process_query.0.filter_by.#", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.3.process_query.0.filter_by.0", "active"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.3.process_query.0.limit", "50"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.request.3.display_type", "area"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.marker.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.marker.0.display_type", "error dashed"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.marker.0.label", " z=6 "),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.marker.0.value", "y = 4"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.marker.1.display_type", "ok solid"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.marker.1.label", " x=8 "),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.marker.1.value", "10 < y < 999"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.show_legend", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.time.live_span", "1h"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.event.0.q", "sources:test tags:1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.event.1.q", "sources:test tags:2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.yaxis.0.scale", "log"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.yaxis.0.include_zero", "false"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.10.timeseries_definition.0.yaxis.0.max", "100"),
					// Toplist widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.11.toplist_definition.0.request.0.q", "avg:system.cpu.user{app:general} by {env}"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.11.toplist_definition.0.request.0.conditional_formats.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.11.toplist_definition.0.request.0.conditional_formats.0.comparator", "<"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.11.toplist_definition.0.request.0.conditional_formats.0.value", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.11.toplist_definition.0.request.0.conditional_formats.0.palette", "white_on_green"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.11.toplist_definition.0.request.0.conditional_formats.1.comparator", ">"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.11.toplist_definition.0.request.0.conditional_formats.1.value", "2.2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.11.toplist_definition.0.request.0.conditional_formats.1.palette", "white_on_red"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.11.toplist_definition.0.title", "Widget Title"),
					// Group widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.layout_type", "ordered"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.title", "Group Widget"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.#", "2"),
					// Inner Note widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.0.note_definition.0.content", "cluster note widget"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.0.note_definition.0.background_color", "yellow"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.0.note_definition.0.font_size", "16"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.0.note_definition.0.text_align", "left"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.0.note_definition.0.show_tick", "false"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.0.note_definition.0.tick_edge", "left"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.0.note_definition.0.tick_pos", "50%"),
					// Inner Alert Graph widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.1.alert_graph_definition.0.alert_id", "123"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.1.alert_graph_definition.0.viz_type", "toplist"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.1.alert_graph_definition.0.title", "Alert Graph"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.12.group_definition.0.widget.1.alert_graph_definition.0.time.live_span", "1h"),
					// Service Level Objective widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.13.scatterplot_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.13.scatterplot_definition.0.view_type", "detail"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.13.scatterplot_definition.0.slo_id", "56789"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.13.scatterplot_definition.0.show_error_budget", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.13.scatterplot_definition.0.view_mode", "overall"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.13.scatterplot_definition.0.time_windows.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.13.scatterplot_definition.0.time_windows.0", "7d"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.13.scatterplot_definition.0.time_windows.1", "previous_week"),
					// Template Variables
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.0.name", "var_1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.0.prefix", "host"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.0.default", "aws"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.1.name", "var_2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.1.prefix", "service_name"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.1.default", "autoscaling"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "description", "Created using the Datadog provider in Terraform"),

					// Free layout dashboard

					// Dashboard metadata
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "title", "Acceptance Test Free Dashboard"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "description", "Created using the Datadog provider in Terraform"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "layout_type", "free"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "is_read_only", "false"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.#", "8"),

					// Event Stream widget
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.event_stream_definition.0.query", "*"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.event_stream_definition.0.event_size", "l"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.event_stream_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.event_stream_definition.0.title_size", "16"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.event_stream_definition.0.title_align", "left"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.event_stream_definition.0.time.live_span", "1h"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.layout.height", "43"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.layout.width", "32"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.layout.x", "5"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.layout.y", "5"),
					// Event Timeline widget
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.1.event_timeline_definition.0.query", "*"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.1.event_timeline_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.1.event_timeline_definition.0.title_align", "left"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.1.event_timeline_definition.0.title_size", "16"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.1.event_timeline_definition.0.time.live_span", "1h"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.1.layout.height", "9"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.1.layout.width", "65"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.1.layout.x", "42"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.1.layout.y", "73"),
					// Free Text widget
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.2.free_text_definition.0.text", "free text content"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.2.free_text_definition.0.color", "#d00"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.2.free_text_definition.0.font_size", "88"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.2.free_text_definition.0.text_align", "left"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.2.layout.height", "20"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.2.layout.width", "30"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.2.layout.x", "42"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.2.layout.y", "5"),
					// Iframe widget
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.3.iframe_definition.0.url", "http://google.com"),
					// Image widget
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.4.image_definition.0.url", "https://images.pexels.com/photos/67636/rose-blue-flower-rose-blooms-67636.jpeg?auto=compress&cs=tinysrgb&h=350"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.4.image_definition.0.sizing", "fit"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.4.image_definition.0.margin", "small"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.4.layout.height", "20"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.4.layout.width", "30"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.4.layout.x", "77"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.4.layout.y", "7"),
					// Log Stream widget
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.log_stream_definition.0.logset", "19"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.log_stream_definition.0.query", "error"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.log_stream_definition.0.columns.#", "3"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.log_stream_definition.0.columns.0", "core_host"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.log_stream_definition.0.columns.1", "core_service"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.log_stream_definition.0.columns.2", "tag_source"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.layout.height", "36"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.layout.width", "32"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.layout.x", "5"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.5.layout.y", "51"),
					// Manage Status widget
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.color_preference", "text"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.count", "50"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.display_format", "countsAndList"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.hide_zero_counts", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.query", "type:metric"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.sort", "status,asc"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.start", "0"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.title_align", "left"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.manage_status_definition.0.title_size", "16"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.layout.height", "40"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.layout.width", "30"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.layout.x", "112"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.6.layout.y", "55"),
					// Trace Service widget
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.display_format", "three_column"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.env", "datad0g.com"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.service", "alerting-cassandra"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.show_breakdown", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.show_distribution", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.show_errors", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.show_hits", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.show_latency", "false"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.show_resource_list", "false"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.size_format", "large"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.span_name", "cassandra.query"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.title", "alerting-cassandra #env:datad0g.com"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.title_align", "center"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.title_size", "13"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.7.trace_service_definition.0.time.live_span", "1h"),
					// Template Variables
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.0.default", "aws"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.0.name", "var_1"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.0.prefix", "host"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.1.default", "autoscaling"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.1.name", "var_2"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.1.prefix", "service_name"),
				),
			},
		},
	})
}

func TestAccDatadogDashboard_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardConfig,
			},
			{
				ResourceName:      "datadog_dashboard.ordered_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "datadog_dashboard.free_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkDashboardExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	for _, r := range s.RootModule().Resources {
		if _, err := client.GetBoard(r.Primary.ID); err != nil {
			return fmt.Errorf("Received an error retrieving dashboard1 %s", err)
		}
	}
	return nil
}

func checkDashboardDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	for _, r := range s.RootModule().Resources {
		if _, err := client.GetBoard(r.Primary.ID); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("Received an error retrieving dashboard2 %s", err)
		}
		return fmt.Errorf("Dashboard still exists")
	}
	return nil
}
