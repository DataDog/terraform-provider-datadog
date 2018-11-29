package datadog

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

const config = `
resource "datadog_screenboard" "acceptance_test" {
	title = "Acceptance Test Screenboard"
	read_only = true
	width = "640"
	height = 480

	template_variable {
		name    = "varname 1"
		prefix  = "pod_name"
		default = "*"
	}

	template_variable {
		name    = "varname 2"
		prefix  = "service_name"
		default = "autoscaling"
	}

	widget {
		type       = "free_text"
		x          = 5
		y          = 5
		text       = "test text"
		text_align = "right"
		font_size  = "36"
		color      = "#ffc0cb"
	}

	widget {
		type        = "timeseries"
		x           = 25
		y           = 5
		title       = "graph title terraform"
		title_size  = 16
		title_align = "right"
		legend      = true
		legend_size = 16

		time {
			live_span = "1d"
		}

		tile_def {
			viz = "timeseries"

			request {
				q    = "avg:system.cpu.user{*}"
				type = "line"

				style {
					palette = "purple"
					type    = "dashed"
					width   = "thin"
				}
			}

			marker {
				label = "test marker"
				type  = "error dashed"
				value = "y < 6"
			}

			event {
				q = "test event"
			}
		}
	}

	widget {
		type        = "query_value"
		x           = 45
		y           = 25
		title       = "query value title terraform"
		title_size  = 20
		title_align = "center"
		legend      = true
		legend_size = 16

		tile_def {
			viz = "query_value"

			request {
				q    = "avg:system.cpu.user{*}"
				type = "line"

				style {
					palette = "purple"
					type    = "dashed"
					width   = "thin"
				}

				conditional_format {
					comparator = ">"
					value      = "1"
					palette    = "white_on_red"
				}

				conditional_format {
					comparator = ">="
					value      = "2"
					palette    = "white_on_yellow"
				}

				aggregator = "max"
			}

			custom_unit = "%"
			autoscale   = false
			precision   = "6"
			text_align  = "right"
		}
	}

	widget {
		type        = "toplist"
		x           = 65
		y           = 5
		title       = "toplist title terraform"
		legend      = true
		legend_size = "auto"

		time {
			live_span = "1d"
		}

		tile_def {
			viz = "toplist"

			request {
				q = "top(avg:system.load.1{*} by {host}, 10, 'mean', 'desc')"

				style {
					palette = "purple"
					type    = "dashed"
					width   = "thin"
				}

				conditional_format {
					comparator = ">"
					value      = "4"
					palette    = "white_on_green"
				}
			}
		}
	}

	widget {
		type  = "change"
		x     = 85
		y     = 5
		title = "change title terraform"

		tile_def {
			viz = "change"

			request {
				q             = "min:system.load.1{*} by {host}"
				compare_to    = "week_before"
				change_type   = "relative"
				order_by      = "present"
				order_dir     = "asc"
				extra_col     = ""
				increase_good = false
			}
		}
	}

	widget {
		type  = "event_timeline"
		x     = 105
		y     = 5
		title = "event_timeline title terraform"
		query = "status:error"

		time {
			live_span = "1d"
		}
	}

	widget {
		type       = "event_stream"
		x          = 115
		y          = 5
		title      = "event_stream title terraform"
		query      = "*"
		event_size = "l"

		time {
			live_span = "4h"
		}
	}

	widget {
		type   = "image"
		x      = 145
		y      = 5
		title  = "image title terraform"
		sizing = "fit"
		margin = "large"
		url    = "https://datadog-prod.imgix.net/img/dd_logo_70x75.png"
	}

	widget {
		type       = "note"
		x          = 165
		y          = 5
		bgcolor    = "pink"
		text_align = "right"
		font_size  = "36"
		tick       = true
		tick_edge  = "bottom"
		tick_pos   = "50%"
		html       = "<b>test note</b>"
	}

	widget {
		type     = "alert_graph"
		x        = 185
		y        = 5
		title    = "alert graph title terraform"
		alert_id = "123456"
		viz_type = "toplist"

		time {
			live_span = "15m"
		}
	}

	widget {
		type       = "alert_value"
		x          = 205
		y          = 5
		title      = "alert value title terraform"
		alert_id   = "123456"
		text_size  = "fill_height"
		text_align = "right"
		precision  = "*"
		unit       = "b"
	}

	widget {
		type = "iframe"
		x    = 225
		y    = 5
		url  = "https://www.datadoghq.org"
	}

	widget {
		type        = "check_status"
		x           = 245
		y           = 5
		title       = "test title"
		title_align = "left"
		grouping    = "check"
		check       = "aws.ecs.agent_connected"
		tags        = ["*"]
		group       = "cluster:test"

		time {
			live_span = "30m"
		}
	}

	widget {
		type                    = "trace_service"
		x                       = 265
		y                       = 5
		env                     = "testEnv"
		service_service         = ""
		service_name            = ""
		size_version            = "large"
		layout_version          = "three_column"
		must_show_hits          = true
		must_show_errors        = true
		must_show_latency       = true
		must_show_breakdown     = true
		must_show_distribution  = true
		must_show_resource_list = true

		time {
			live_span = "30m"
		}
	}

	widget {
		type  = "hostmap"
		x     = 285
		y     = 5
		query = "avg:system.load.1{*} by {host}"

		tile_def {
			viz             = "hostmap"
			node_type       = "container"
			scope           = ["datacenter:test"]
			group           = ["pod_name"]
			no_group_hosts  = false
			no_metric_hosts = false

			request {
				q    = "max:process.stat.container.io.wbps{datacenter:test} by {host}"
				type = "fill"
			}

			style {
				palette      = "hostmap_blues"
				palette_flip = true
				fill_min     = 20
				fill_max     = 300
			}
		}
	}

	widget {
		type                      = "manage_status"
		x                         = 305
		y                         = 5
		display_format            = "countsAndList"
		color_preference          = "background"
		hide_zero_counts          = true
		manage_status_show_title  = false
		manage_status_title_text  = "test title"
		manage_status_title_size  = "20"
		manage_status_title_align = "right"

		params {
			sort  = "status,asc"
			text  = "status:alert"
			count = 50
			start = 0
		}
	}

	widget {
		type    = "log_stream"
		x       = 325
		y       = 5
		query   = "source:kubernetes"
		columns = "[\"column1\",\"column2\",\"column3\"]"
		logset  = "1234"

		time {
			live_span = "1h"
		}
	}

	widget {
		type = "process"
		x    = 365
		y    = 5

		tile_def {
			viz = "process"

			request {
				query_type  = "process"
				metric      = "process.stat.cpu.total_pct"
				text_filter = ""
				tag_filters = []
				limit       = 200

				style = {
					palette = "dog_classic_area"
				}
			}
		}
	}
}
`

func TestAccDatadogScreenboard_update(t *testing.T) {

	step1 := resource.TestStep{
		Config: config,
		Check: resource.ComposeTestCheckFunc(
			checkScreenboardExists,
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "width", "640"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "height", "480"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "template_variable.#", "2"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "template_variable.0.default", "*"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "template_variable.0.name", "varname 1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "template_variable.0.prefix", "pod_name"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "template_variable.1.default", "autoscaling"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "template_variable.1.name", "varname 2"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "template_variable.1.prefix", "service_name"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.#", "18"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.color", "#ffc0cb"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.font_size", "36"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.text", "test text"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.text_align", "right"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.title", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.type", "free_text"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.x", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.0.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.legend", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.legend_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.autoscale", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.custom_unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.event.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.event.0.q", "test event"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.group.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.marker.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.marker.0.label", "test marker"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.marker.0.type", "error dashed"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.marker.0.value", "y < 6"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.no_group_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.no_metric_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.node_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.aggregator", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.change_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.compare_to", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.conditional_format.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.extra_col", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.increase_good", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.limit", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.metric", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.order_by", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.order_dir", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.q", "avg:system.cpu.user{*}"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.query_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.style.%", "3"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.style.palette", "purple"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.style.type", "dashed"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.style.width", "thin"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.tag_filters.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.text_filter", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.request.0.type", "line"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.scope.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.style.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.tile_def.0.viz", "timeseries"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.time.%", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.time.live_span", "1d"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.title", "graph title terraform"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.title_align", "right"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.type", "timeseries"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.x", "25"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.1.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.alert_id", "123456"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.precision", "*"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.text_align", "right"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.text_size", "fill_height"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.title", "alert value title terraform"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.type", "alert_value"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.unit", "b"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.x", "205"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.10.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.title", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.type", "iframe"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.url", "https://www.datadoghq.org"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.x", "225"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.11.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.check", "aws.ecs.agent_connected"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.group", "cluster:test"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.grouping", "check"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.tags.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.tags.0", "*"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.time.%", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.time.live_span", "30m"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.title", "test title"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.type", "check_status"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.x", "245"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.12.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.env", "testEnv"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.layout_version", "three_column"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.must_show_breakdown", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.must_show_distribution", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.must_show_errors", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.must_show_hits", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.must_show_latency", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.must_show_resource_list", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.size_version", "large"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.time.%", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.time.live_span", "30m"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.title", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.type", "trace_service"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.x", "265"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.13.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.query", "avg:system.load.1{*} by {host}"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.autoscale", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.custom_unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.event.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.group.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.group.0", "pod_name"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.marker.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.no_group_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.no_metric_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.node_type", "container"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.aggregator", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.change_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.compare_to", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.conditional_format.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.extra_col", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.increase_good", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.limit", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.metric", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.order_by", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.order_dir", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.q", "max:process.stat.container.io.wbps{datacenter:test} by {host}"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.query_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.style.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.tag_filters.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.text_filter", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.request.0.type", "fill"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.scope.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.scope.0", "datacenter:test"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.style.%", "4"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.style.fill_max", "300"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.style.fill_min", "20"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.style.palette", "hostmap_blues"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.style.palette_flip", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.tile_def.0.viz", "hostmap"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.title", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.type", "hostmap"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.x", "285"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.14.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.color_preference", "background"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.display_format", "countsAndList"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.hide_zero_counts", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.manage_status_title_align", "right"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.manage_status_title_size", "20"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.manage_status_title_text", "test title"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.params.%", "4"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.params.count", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.params.sort", "status,asc"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.params.start", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.params.text", "status:alert"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.title", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.type", "manage_status"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.x", "305"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.15.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.columns", "[\"column1\",\"column2\",\"column3\"]"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.logset", "1234"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.query", "source:kubernetes"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.time.%", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.time.live_span", "1h"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.title", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.type", "log_stream"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.x", "325"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.16.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.autoscale", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.custom_unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.event.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.group.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.marker.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.no_group_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.no_metric_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.node_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.aggregator", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.change_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.compare_to", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.conditional_format.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.extra_col", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.increase_good", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.limit", "200"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.metric", "process.stat.cpu.total_pct"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.order_by", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.order_dir", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.q", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.query_type", "process"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.style.%", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.style.palette", "dog_classic_area"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.tag_filters.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.text_filter", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.request.0.type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.scope.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.style.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.tile_def.0.viz", "process"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.title", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.type", "process"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.x", "365"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.17.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.legend", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.legend_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.autoscale", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.custom_unit", "%"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.event.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.group.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.marker.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.no_group_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.no_metric_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.node_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.precision", "6"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.aggregator", "max"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.change_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.compare_to", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.#", "2"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.0.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.0.comparator", ">"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.0.invert", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.0.palette", "white_on_red"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.0.value", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.1.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.1.comparator", ">="),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.1.invert", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.1.palette", "white_on_yellow"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.conditional_format.1.value", "2"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.extra_col", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.increase_good", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.limit", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.metric", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.order_by", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.order_dir", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.q", "avg:system.cpu.user{*}"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.query_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.style.%", "3"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.style.palette", "purple"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.style.type", "dashed"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.style.width", "thin"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.tag_filters.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.text_filter", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.request.0.type", "line"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.scope.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.style.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.text_align", "right"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.tile_def.0.viz", "query_value"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.title", "query value title terraform"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.title_align", "center"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.title_size", "20"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.type", "query_value"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.x", "45"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.2.y", "25"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.legend", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.legend_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.autoscale", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.custom_unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.event.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.group.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.marker.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.no_group_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.no_metric_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.node_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.aggregator", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.change_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.compare_to", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.conditional_format.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.conditional_format.0.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.conditional_format.0.comparator", ">"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.conditional_format.0.invert", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.conditional_format.0.palette", "white_on_green"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.conditional_format.0.value", "4"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.extra_col", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.increase_good", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.limit", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.metric", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.order_by", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.order_dir", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.q", "top(avg:system.load.1{*} by {host}, 10, 'mean', 'desc')"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.query_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.style.%", "3"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.style.palette", "purple"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.style.type", "dashed"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.style.width", "thin"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.tag_filters.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.text_filter", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.request.0.type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.scope.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.style.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.tile_def.0.viz", "toplist"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.time.%", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.time.live_span", "1d"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.title", "toplist title terraform"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.type", "toplist"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.x", "65"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.3.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.autoscale", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.custom_unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.event.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.group.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.marker.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.no_group_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.no_metric_hosts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.node_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.#", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.aggregator", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.change_type", "relative"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.compare_to", "week_before"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.conditional_format.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.extra_col", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.increase_good", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.limit", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.metric", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.order_by", "present"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.order_dir", "asc"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.q", "min:system.load.1{*} by {host}"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.query_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.style.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.tag_filters.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.text_filter", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.request.0.type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.scope.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.style.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.tile_def.0.viz", "change"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.title", "change title terraform"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.type", "change"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.x", "85"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.4.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.query", "status:error"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.time.%", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.time.live_span", "1d"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.title", "event_timeline title terraform"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.type", "event_timeline"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.x", "105"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.5.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.event_size", "l"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.query", "*"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.time.%", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.time.live_span", "4h"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.title", "event_stream title terraform"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.type", "event_stream"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.x", "115"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.6.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.margin", "large"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.sizing", "fit"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.title", "image title terraform"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.type", "image"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.url", "https://datadog-prod.imgix.net/img/dd_logo_70x75.png"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.x", "145"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.7.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.alert_id", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.bgcolor", "pink"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.font_size", "36"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.html", "<b>test note</b>"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.text_align", "right"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.tick", "true"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.tick_edge", "bottom"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.tick_pos", "50%"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.time.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.title", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.type", "note"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.viz_type", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.x", "165"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.8.y", "5"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.alert_id", "123456"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.auto_refresh", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.bgcolor", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.check", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.color", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.color_preference", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.columns", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.display_format", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.env", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.event_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.font_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.group", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.group_by.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.grouping", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.height", "15"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.hide_zero_counts", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.html", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.layout_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.legend", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.legend_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.logset", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.manage_status_show_title", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.manage_status_title_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.manage_status_title_size", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.manage_status_title_text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.margin", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.monitor.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.must_show_breakdown", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.must_show_distribution", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.must_show_errors", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.must_show_hits", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.must_show_latency", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.must_show_resource_list", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.params.%", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.precision", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.query", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.rule.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.service_name", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.service_service", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.size_version", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.sizing", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.tags.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.text", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.text_align", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.text_size", "auto"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.tick", "false"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.tick_edge", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.tick_pos", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.tile_def.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.time.%", "1"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.time.live_span", "15m"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.timeframes.#", "0"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.title", "alert graph title terraform"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.title_align", "left"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.title_size", "16"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.type", "alert_graph"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.unit", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.url", ""),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.viz_type", "toplist"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.width", "50"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.x", "185"),
			resource.TestCheckResourceAttr("datadog_screenboard.acceptance_test", "widget.9.y", "5"),
		),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkScreenboardDestroy,
		Steps:        []resource.TestStep{step1},
	})
}

func checkScreenboardExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	for _, r := range s.RootModule().Resources {
		i, _ := strconv.Atoi(r.Primary.ID)
		if _, err := client.GetScreenboard(i); err != nil {
			return fmt.Errorf("Received an error retrieving screenboard %s", err)
		}
	}
	return nil
}

func checkScreenboardDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	for _, r := range s.RootModule().Resources {
		i, _ := strconv.Atoi(r.Primary.ID)
		if _, err := client.GetScreenboard(i); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("Received an error retrieving screenboard %s", err)
		}
		return fmt.Errorf("Screenboard still exists")
	}
	return nil
}
