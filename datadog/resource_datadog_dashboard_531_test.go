package datadog

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Unit test to reproduce https://github.com/terraform-providers/terraform-provider-datadog/issues/531
// the only change made was to change:
//template_variable {
//	name    = "service"
//	prefix  = "service/component"
//	default = var.service
//}
// into
//template_variable {
//	name    = "service"
//	prefix  = "service/component"
//	default = "service"
//}
const datadogIssue531Config = `
resource "datadog_dashboard" "apache" {
  title        = "Apache"
  layout_type  = "free"
  is_read_only = true

  template_variable {
    name    = "service"
    prefix  = "service/component"
    default = "service"
  }

  template_variable {
    name    = "region"
    prefix  = "region"
    default = "*"
  }

  template_variable {
    name    = "environment"
    prefix  = "environment"
    default = "*"
  }

  widget {
    image_definition {
      url    = "--redacted--"
      sizing = "zoom"
    }
    layout = {
      x      = 1
      y      = 1
      width  = 17
      height = 11
    }
  }

  widget {
    image_definition {
      url    = "--redacted--"
      sizing = "center"
    }
    layout = {
      x      = 1
      y      = 13
      width  = 17
      height = 12
    }
  }

  widget {
    note_definition {
      content          = "**Availability**"
      text_align       = "center"
      font_size        = "18"
      tick_pos         = "50%"
      show_tick        = true
      tick_edge        = "bottom"
      background_color = "gray"
    }
    layout = {
      x      = 1
      y      = 26
      width  = 17
      height = 5
    }
  }

  widget {
    check_status_definition {
      title       = "Server can connect"
      title_align = "center"
      title_size  = 16

      check    = "apache.can_connect"
      grouping = "cluster"

      tags = [
        "$service",
        "$region",
        "$environment",
      ]

      time = {
        live_span = "10m"
      }
    }
    layout = {
      x      = 1
      y      = 33
      width  = 17
      height = 13
    }
  }

  widget {
    query_value_definition {
      title       = "Min Uptime"
      title_align = "left"
      title_size  = 16

      autoscale = true

      request {
        q          = "min:apache.performance.uptime{$service,$region,$environment}"
        aggregator = "last"

        conditional_formats {
          palette    = "white_on_green"
          value      = 0
          comparator = ">"
        }
      }
    }
    layout = {
      x      = 1
      y      = 48
      width  = 17
      height = 13
    }
  }

  widget {
    note_definition {
      content          = "**Throughput**"
      text_align       = "center"
      font_size        = "18"
      tick_pos         = "50%"
      show_tick        = true
      tick_edge        = "bottom"
      background_color = "gray"
    }
    layout = {
      x      = 20
      y      = 1
      width  = 47
      height = 5
    }
  }

  widget {
    timeseries_definition {
      title       = "Rate of requests (per region)"
      title_align = "left"
      title_size  = 16

      request {
        q            = "sum:apache.net.request_per_s{$service,$region,$environment} by {region}"
        display_type = "line"

        style {
          palette    = "cool"
          line_type  = "solid"
          line_width = "normal"
        }
      }
    }
    layout = {
      x      = 20
      y      = 8
      width  = 47
      height = 17
    }
  }

  widget {
    timeseries_definition {
      title       = "Bytes served (per region)"
      title_align = "left"
      title_size  = 16

      request {
        q            = "sum:apache.net.bytes_per_s{$service,$region,$environment} by {region}"
        display_type = "line"

        style {
          palette    = "cool"
          line_type  = "solid"
          line_width = "normal"
        }
      }
    }
    layout = {
      x      = 20
      y      = 26
      width  = 47
      height = 17
    }
  }

  widget {
    timeseries_definition {
      title       = "Requests per second (per region)"
      title_align = "left"
      title_size  = 16

      request {
        q            = "avg:apache.net.request_per_s{$service,$region,$environment} by {region}"
        display_type = "line"

        style {
          palette    = "cool"
          line_type  = "solid"
          line_width = "normal"
        }
      }
    }
    layout = {
      x      = 20
      y      = 44
      width  = 47
      height = 17
    }
  }

  widget {
    note_definition {
      content          = "**Resource Utilisation**"
      text_align       = "center"
      font_size        = "18"
      tick_pos         = "50%"
      show_tick        = true
      tick_edge        = "bottom"
      background_color = "gray"
    }
    layout = {
      x      = 69
      y      = 1
      width  = 53
      height = 5
    }
  }

  widget {
    toplist_definition {
      title       = "Apache process CPU usage (top 10 containers)"
      title_align = "left"
      title_size  = 16

      request {
        q = "top(avg:apache.performance.cpu_load{$service,$region,$environment,short_image:shibboleth} by {region,image_tag,ecs_container_name}, 10, 'mean', 'desc')"

        style {
          palette = "dog_classic"
        }
      }
    }
    layout = {
      x      = 69
      y      = 8
      width  = 53
      height = 32
    }
  }

  widget {
    timeseries_definition {
      title       = "Busy vs. idle worker threads"
      title_align = "left"
      title_size  = 16

      request {
        q            = "sum:apache.performance.idle_workers{$service,$region,$environment}, sum:apache.performance.busy_workers{$service,$region,$environment}"
        display_type = "area"

        style {
          palette    = "cool"
          line_type  = "solid"
          line_width = "normal"
        }
      }
    }
    layout = {
      x      = 69
      y      = 41
      width  = 53
      height = 20
    }
  }

  widget {
    note_definition {
      content          = "**Connection Status**"
      text_align       = "center"
      font_size        = "18"
      tick_pos         = "50%"
      show_tick        = true
      tick_edge        = "bottom"
      background_color = "gray"
    }
    layout = {
      x      = 124
      y      = 1
      width  = 47
      height = 5
    }
  }

  widget {
    timeseries_definition {
      title       = "Async connections: writing, keep-alive, closing"
      title_align = "left"
      title_size  = 16

      request {
        q            = "sum:apache.conns_async_closing{$service,$region,$environment}.rollup(max), sum:apache.conns_async_writing{$service,$region,$environment}.rollup(max), sum:apache.conns_async_keep_alive{$service,$region,$environment}.rollup(max)"
        display_type = "bars"

        style {
          palette    = "cool"
          line_type  = "solid"
          line_width = "normal"
        }
      }
    }
    layout = {
      x      = 124
      y      = 8
      width  = 47
      height = 17
    }
  }

  widget {
    timeseries_definition {
      title       = "Total async connections"
      title_align = "left"
      title_size  = 16

      request {
        q            = "sum:apache.conns_total{$service,$region,$environment} by {region}.rollup(max)"
        display_type = "bars"

        style {
          palette    = "cool"
          line_type  = "solid"
          line_width = "normal"
        }
      }
    }
    layout = {
      x      = 124
      y      = 26
      width  = 47
      height = 17
    }
  }
}
`

var datadogIssue531Asserts = []string{
	"is_read_only = true",
	"layout_type = free",
	"notify_list.# = 0",
	"template_variable.# = 3",
	"template_variable.0.default = service",
	"template_variable.0.name = service",
	"template_variable.0.prefix = service/component",
	"template_variable.1.default = *",
	"template_variable.1.name = region",
	"template_variable.1.prefix = region",
	"template_variable.2.default = *",
	"template_variable.2.name = environment",
	"template_variable.2.prefix = environment",
	"template_variable_preset.# = 0",
	"title = Apache",
	"widget.# = 15",
	"widget.0.image_definition.0.margin =",
	"widget.0.image_definition.0.sizing = zoom",
	"widget.0.image_definition.0.url = --redacted--",
	"widget.0.layout.height = 11",
	"widget.0.layout.width = 17",
	"widget.0.layout.x = 1",
	"widget.0.layout.y = 1",
	"widget.1.image_definition.0.margin =",
	"widget.1.image_definition.0.sizing = center",
	"widget.1.image_definition.0.url = --redacted--",
	"widget.1.layout.height = 12",
	"widget.1.layout.width = 17",
	"widget.1.layout.x = 1",
	"widget.1.layout.y = 13",
	"widget.10.layout.height = 32",
	"widget.10.layout.width = 53",
	"widget.10.layout.x = 69",
	"widget.10.layout.y = 8",
	"widget.10.toplist_definition.0.request.# = 1",
	"widget.10.toplist_definition.0.request.0.conditional_formats.# = 0",
	"widget.10.toplist_definition.0.request.0.q = top(avg:apache.performance.cpu_load{$service,$region,$environment,short_image:shibboleth} by {region,image_tag,ecs_container_name}, 10, 'mean', 'desc')",
	"widget.10.toplist_definition.0.request.0.style.# = 1",
	"widget.10.toplist_definition.0.request.0.style.0.palette = dog_classic",
	"widget.10.toplist_definition.0.title = Apache process CPU usage (top 10 containers)",
	"widget.10.toplist_definition.0.title_align = left",
	"widget.10.toplist_definition.0.title_size = 16",
	"widget.11.layout.height = 20",
	"widget.11.layout.width = 53",
	"widget.11.layout.x = 69",
	"widget.11.layout.y = 41",
	"widget.11.timeseries_definition.0.event.# = 0",
	"widget.11.timeseries_definition.0.legend_size =",
	"widget.11.timeseries_definition.0.marker.# = 0",
	"widget.11.timeseries_definition.0.request.# = 1",
	"widget.11.timeseries_definition.0.request.0.display_type = area",
	"widget.11.timeseries_definition.0.request.0.metadata.# = 0",
	"widget.11.timeseries_definition.0.request.0.q = sum:apache.performance.idle_workers{$service,$region,$environment}, sum:apache.performance.busy_workers{$service,$region,$environment}",
	"widget.11.timeseries_definition.0.request.0.style.# = 1",
	"widget.11.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.11.timeseries_definition.0.request.0.style.0.line_width = normal",
	"widget.11.timeseries_definition.0.request.0.style.0.palette = cool",
	"widget.11.timeseries_definition.0.show_legend = false",
	"widget.11.timeseries_definition.0.title = Busy vs. idle worker threads",
	"widget.11.timeseries_definition.0.title_align = left",
	"widget.11.timeseries_definition.0.title_size = 16",
	"widget.11.timeseries_definition.0.yaxis.# = 0",
	"widget.12.layout.height = 5",
	"widget.12.layout.width = 47",
	"widget.12.layout.x = 124",
	"widget.12.layout.y = 1",
	"widget.12.note_definition.0.background_color = gray",
	"widget.12.note_definition.0.content = **Connection Status**",
	"widget.12.note_definition.0.font_size = 18",
	"widget.12.note_definition.0.show_tick = true",
	"widget.12.note_definition.0.text_align = center",
	"widget.12.note_definition.0.tick_edge = bottom",
	"widget.12.note_definition.0.tick_pos = 50%",
	"widget.13.layout.height = 17",
	"widget.13.layout.width = 47",
	"widget.13.layout.x = 124",
	"widget.13.layout.y = 8",
	"widget.13.timeseries_definition.0.event.# = 0",
	"widget.13.timeseries_definition.0.legend_size =",
	"widget.13.timeseries_definition.0.marker.# = 0",
	"widget.13.timeseries_definition.0.request.# = 1",
	"widget.13.timeseries_definition.0.request.0.display_type = bars",
	"widget.13.timeseries_definition.0.request.0.metadata.# = 0",
	"widget.13.timeseries_definition.0.request.0.q = sum:apache.conns_async_closing{$service,$region,$environment}.rollup(max), sum:apache.conns_async_writing{$service,$region,$environment}.rollup(max), sum:apache.conns_async_keep_alive{$service,$region,$environment}.rollup(max)",
	"widget.13.timeseries_definition.0.request.0.style.# = 1",
	"widget.13.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.13.timeseries_definition.0.request.0.style.0.line_width = normal",
	"widget.13.timeseries_definition.0.request.0.style.0.palette = cool",
	"widget.13.timeseries_definition.0.show_legend = false",
	"widget.13.timeseries_definition.0.title = Async connections: writing, keep-alive, closing",
	"widget.13.timeseries_definition.0.title_align = left",
	"widget.13.timeseries_definition.0.title_size = 16",
	"widget.13.timeseries_definition.0.yaxis.# = 0",
	"widget.14.layout.height = 17",
	"widget.14.layout.width = 47",
	"widget.14.layout.x = 124",
	"widget.14.layout.y = 26",
	"widget.14.timeseries_definition.0.event.# = 0",
	"widget.14.timeseries_definition.0.legend_size =",
	"widget.14.timeseries_definition.0.marker.# = 0",
	"widget.14.timeseries_definition.0.request.# = 1",
	"widget.14.timeseries_definition.0.request.0.display_type = bars",
	"widget.14.timeseries_definition.0.request.0.metadata.# = 0",
	"widget.14.timeseries_definition.0.request.0.q = sum:apache.conns_total{$service,$region,$environment} by {region}.rollup(max)",
	"widget.14.timeseries_definition.0.request.0.style.# = 1",
	"widget.14.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.14.timeseries_definition.0.request.0.style.0.line_width = normal",
	"widget.14.timeseries_definition.0.request.0.style.0.palette = cool",
	"widget.14.timeseries_definition.0.show_legend = false",
	"widget.14.timeseries_definition.0.title = Total async connections",
	"widget.14.timeseries_definition.0.title_align = left",
	"widget.14.timeseries_definition.0.title_size = 16",
	"widget.14.timeseries_definition.0.yaxis.# = 0",
	"widget.2.layout.height = 5",
	"widget.2.layout.width = 17",
	"widget.2.layout.x = 1",
	"widget.2.layout.y = 26",
	"widget.2.note_definition.0.background_color = gray",
	"widget.2.note_definition.0.content = **Availability**",
	"widget.2.note_definition.0.font_size = 18",
	"widget.2.note_definition.0.show_tick = true",
	"widget.2.note_definition.0.text_align = center",
	"widget.2.note_definition.0.tick_edge = bottom",
	"widget.2.note_definition.0.tick_pos = 50%",
	"widget.3.check_status_definition.0.check = apache.can_connect",
	"widget.3.check_status_definition.0.group =",
	"widget.3.check_status_definition.0.group_by.# = 0",
	"widget.3.check_status_definition.0.grouping = cluster",
	"widget.3.check_status_definition.0.tags.# = 3",
	"widget.3.check_status_definition.0.tags.0 = $service",
	"widget.3.check_status_definition.0.tags.1 = $region",
	"widget.3.check_status_definition.0.tags.2 = $environment",
	"widget.3.check_status_definition.0.time.live_span = 10m",
	"widget.3.check_status_definition.0.title = Server can connect",
	"widget.3.check_status_definition.0.title_align = center",
	"widget.3.check_status_definition.0.title_size = 16",
	"widget.3.layout.height = 13",
	"widget.3.layout.width = 17",
	"widget.3.layout.x = 1",
	"widget.3.layout.y = 33",
	"widget.4.layout.height = 13",
	"widget.4.layout.width = 17",
	"widget.4.layout.x = 1",
	"widget.4.layout.y = 48",
	"widget.4.query_value_definition.0.autoscale = true",
	"widget.4.query_value_definition.0.custom_unit =",
	"widget.4.query_value_definition.0.precision = 0",
	"widget.4.query_value_definition.0.request.# = 1",
	"widget.4.query_value_definition.0.request.0.aggregator = last",
	"widget.4.query_value_definition.0.request.0.conditional_formats.# = 1",
	"widget.4.query_value_definition.0.request.0.conditional_formats.0.comparator = >",
	"widget.4.query_value_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.4.query_value_definition.0.request.0.conditional_formats.0.custom_fg_color =",
	"widget.4.query_value_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.4.query_value_definition.0.request.0.conditional_formats.0.image_url =",
	"widget.4.query_value_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.4.query_value_definition.0.request.0.conditional_formats.0.timeframe =",
	"widget.4.query_value_definition.0.request.0.conditional_formats.0.value = 0",
	"widget.4.query_value_definition.0.request.0.q = min:apache.performance.uptime{$service,$region,$environment}",
	"widget.4.query_value_definition.0.text_align =",
	"widget.4.query_value_definition.0.title = Min Uptime",
	"widget.4.query_value_definition.0.title_align = left",
	"widget.4.query_value_definition.0.title_size = 16",
	"widget.5.layout.height = 5",
	"widget.5.layout.width = 47",
	"widget.5.layout.x = 20",
	"widget.5.layout.y = 1",
	"widget.5.note_definition.0.background_color = gray",
	"widget.5.note_definition.0.content = **Throughput**",
	"widget.5.note_definition.0.font_size = 18",
	"widget.5.note_definition.0.show_tick = true",
	"widget.5.note_definition.0.text_align = center",
	"widget.5.note_definition.0.tick_edge = bottom",
	"widget.5.note_definition.0.tick_pos = 50%",
	"widget.6.layout.height = 17",
	"widget.6.layout.width = 47",
	"widget.6.layout.x = 20",
	"widget.6.layout.y = 8",
	"widget.6.timeseries_definition.0.event.# = 0",
	"widget.6.timeseries_definition.0.legend_size =",
	"widget.6.timeseries_definition.0.marker.# = 0",
	"widget.6.timeseries_definition.0.request.# = 1",
	"widget.6.timeseries_definition.0.request.0.display_type = line",
	"widget.6.timeseries_definition.0.request.0.metadata.# = 0",
	"widget.6.timeseries_definition.0.request.0.q = sum:apache.net.request_per_s{$service,$region,$environment} by {region}",
	"widget.6.timeseries_definition.0.request.0.style.# = 1",
	"widget.6.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.6.timeseries_definition.0.request.0.style.0.line_width = normal",
	"widget.6.timeseries_definition.0.request.0.style.0.palette = cool",
	"widget.6.timeseries_definition.0.show_legend = false",
	"widget.6.timeseries_definition.0.title = Rate of requests (per region)",
	"widget.6.timeseries_definition.0.title_align = left",
	"widget.6.timeseries_definition.0.title_size = 16",
	"widget.6.timeseries_definition.0.yaxis.# = 0",
	"widget.7.layout.height = 17",
	"widget.7.layout.width = 47",
	"widget.7.layout.x = 20",
	"widget.7.layout.y = 26",
	"widget.7.timeseries_definition.0.event.# = 0",
	"widget.7.timeseries_definition.0.legend_size =",
	"widget.7.timeseries_definition.0.marker.# = 0",
	"widget.7.timeseries_definition.0.request.# = 1",
	"widget.7.timeseries_definition.0.request.0.display_type = line",
	"widget.7.timeseries_definition.0.request.0.metadata.# = 0",
	"widget.7.timeseries_definition.0.request.0.q = sum:apache.net.bytes_per_s{$service,$region,$environment} by {region}",
	"widget.7.timeseries_definition.0.request.0.style.# = 1",
	"widget.7.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.7.timeseries_definition.0.request.0.style.0.line_width = normal",
	"widget.7.timeseries_definition.0.request.0.style.0.palette = cool",
	"widget.7.timeseries_definition.0.show_legend = false",
	"widget.7.timeseries_definition.0.title = Bytes served (per region)",
	"widget.7.timeseries_definition.0.title_align = left",
	"widget.7.timeseries_definition.0.title_size = 16",
	"widget.7.timeseries_definition.0.yaxis.# = 0",
	"widget.8.layout.height = 17",
	"widget.8.layout.width = 47",
	"widget.8.layout.x = 20",
	"widget.8.layout.y = 44",
	"widget.8.timeseries_definition.0.event.# = 0",
	"widget.8.timeseries_definition.0.legend_size =",
	"widget.8.timeseries_definition.0.marker.# = 0",
	"widget.8.timeseries_definition.0.request.# = 1",
	"widget.8.timeseries_definition.0.request.0.display_type = line",
	"widget.8.timeseries_definition.0.request.0.metadata.# = 0",
	"widget.8.timeseries_definition.0.request.0.q = avg:apache.net.request_per_s{$service,$region,$environment} by {region}",
	"widget.8.timeseries_definition.0.request.0.style.# = 1",
	"widget.8.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.8.timeseries_definition.0.request.0.style.0.line_width = normal",
	"widget.8.timeseries_definition.0.request.0.style.0.palette = cool",
	"widget.8.timeseries_definition.0.show_legend = false",
	"widget.8.timeseries_definition.0.title = Requests per second (per region)",
	"widget.8.timeseries_definition.0.title_align = left",
	"widget.8.timeseries_definition.0.title_size = 16",
	"widget.8.timeseries_definition.0.yaxis.# = 0",
	"widget.9.layout.height = 5",
	"widget.9.layout.width = 53",
	"widget.9.layout.x = 69",
	"widget.9.layout.y = 1",
	"widget.9.note_definition.0.background_color = gray",
	"widget.9.note_definition.0.content = **Resource Utilisation**",
	"widget.9.note_definition.0.font_size = 18",
	"widget.9.note_definition.0.show_tick = true",
	"widget.9.note_definition.0.text_align = center",
	"widget.9.note_definition.0.tick_edge = bottom",
	"widget.9.note_definition.0.tick_pos = 50%",
}

func testCheckResourceAttrs(name string, checkExists resource.TestCheckFunc, assertions []string) []resource.TestCheckFunc {
	funcs := []resource.TestCheckFunc{}
	funcs = append(funcs, checkExists)
	for _, assertion := range assertions {
		assertionPair := strings.Split(assertion, " = ")
		if len(assertionPair) == 1 {
			assertionPair = strings.Split(assertion, " =")
		}
		key := assertionPair[0]
		value := ""
		if len(assertionPair) > 1 {
			value = assertionPair[1]
		}
		funcs = append(funcs, resource.TestCheckResourceAttr(name, key, value))
	}
	return funcs
}

func TestAccDatadogIssue531(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogIssue531Config,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.apache", checkDashboardExists(accProvider), datadogIssue531Asserts)...,
				),
			},
		},
	})
}
