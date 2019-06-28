---
layout: "datadog"
page_title: "Datadog: datadog_screenboard"
sidebar_current: "docs-datadog-resource-screenboard"
description: |-
  Provides a Datadog screenboard resource. This can be used to create and manage screenboards.
---

# datadog_screenboard

**Note:** This resource is outdated. Use the new [`datadog_dashboard`](dashboard.html) resource instead.

## Example Usage

```hcl
# Create a new Datadog screenboard
resource "datadog_screenboard" "acceptance_test" {
  title     = "Test Screenboard"
  read_only = true

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

    time = {
      live_span = "1d"
    }

    tile_def {
      viz = "timeseries"

      request {
        q    = "avg:system.cpu.user{*}"
        type = "line"

        style = {
          palette = "purple"
          type    = "dashed"
          width   = "thin"
        }

        # NOTE: this will only work with TF >= 0.12; see metadata_json
        # documentation below for example on usage with TF < 0.12
        metadata_json = jsonencode({
          "avg:system.cpu.user{*}": {
            "alias": "CPU Usage"
          }
        })
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

        style = {
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

    time = {
      live_span = "1d"
    }

    tile_def {
      viz = "toplist"

      request {
        q = "top(avg:system.load.1{*} by {host}, 10, 'mean', 'desc')"

        style = {
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

    time = {
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

    time = {
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

    time = {
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

    time = {
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

    time = {
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

      style = {
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

    params = {
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

    time = {
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
```

## Argument Reference

The following arguments are supported:

- `title` - (Required) The name of the screenboard.
- `height` - (Optional) The screenboard's height.
- `width` - (Optional) The screenboard's width.
- `read_only` - (Optional) The read-only status of the screenboard. Default is false.
- `shared` - (Optional) Whether the screenboard is shared or not. Default is false.
- `widget` - (Required) Nested block describing a widget. The structure of this block is described below. Multiple widget blocks are allowed within a datadog_screenboard resource.
- `template_variable` - (Optional) Nested block describing a template variable. The structure of this block is described below. Multiple template_variable blocks are allowed within a datadog_screenboard resource.

### Nested `widget` blocks

Nested `widget` blocks have the following structure:

- `type` - (Required) The type of the widget. One of "free_text", "timeseries", "query_value", "toplist", "change", "event_timeline", "event_stream", "image", "note", "alert_graph", "alert_value", "iframe", "check_status", "trace_service", "hostmap", "manage_status", "log_stream", or "process".
- `x` - (Required) The position of the widget on the x (vertical) axis. Should be greater or equal to 0.
- `y` - (Required) The position of the widget on the y (horizontal) axis. Should be greater or equal to 0.
- `title` - (Optional) The title of the widget.
- `title_align` - (Optional) The alignment of the widget's title. One of "left", "center", or "right".
- `title_size` - (Optional) The size of the widget's title. Default is 16.
- `height` - (Optional) The height of the widget. Default is 15.
- `width` - (Optional) The width of the widget. Default is 50.
- `text` - (Optional, only for widgets of type "free_text") The text to display in the widget.
- `color` - (Optional, only for widgets of type "free_text") The color of the text in the widget.
- `font_size` - (Optional, only for widgets of type "free_text", "note") The size of the text in the widget.
- `text_size` - (Optional, only for widgets of type "alert_value") The size of the text in the widget.
- `unit` - (Optional, only for widgets of type "alert_value") The unit for the value displayed in the widget.
- `precision` - (Optional, only for widgets of type "alert_value") The precision to use when displaying the value. Use "\*" for maximum precision.
- `text_align` - (Optional, only for widgets of type "free_text", "alert_value", "note") The alignment of the text in the widget.
- `alert_id` - (Optional, only for widgets of type "alert_value", "alert_graph") The ID of the monitor used by the widget.
- `auto_refresh` - (Optional, only for widgets of type "alert_value", "alert_graph") Boolean indicating whether the widget is refreshed automatically.
- `legend` - (Optional, only for widgets of type "timeseries", "query_value", "toplist") Boolean indicating whether to display a legend in the widget.
- `legend_size` - (Optional, only for widgets of type "timeseries", "query_value", "toplist") The size of the legend displayed in the widget.
- `query` - (Optional, only for widgets of type "event_timeline", "event_stream", "hostmap", "log_stream") The query to use in the widget.
- `url` - (Optional, only for widgets of type "image", "iframe") The URL to use as a data source for the widget.
- `viz_type` - (Optional, only for widgets of type "alert_graph") Type of visualization to use when displaying the widget. Either "timeseries" or "toplist".
- `tags` - (Optional, only for widgets of type "check_status") List of tags to use in the widget.
- `check` - (Optional, only for widgets of type "check_status") The check to use in the widget.
- `group` - (Optional, only for widgets of type "check_status") The check group to use in the widget.
- `grouping` - (Optional, only for widgets of type "check_status") Either "check" or "cluster", depending on whether the widget should use a single check or a cluster of checks.
- `group_by` - (Optional, only for widgets of type "check_status") When grouping = "cluster", indicates a list of tags to use for grouping.
- `tick` - (Optional, only for widgets of type "note") Boolean indicating whether a tick should be displayed on the border of the widget.
- `tick_pos` - (Optional, only for widgets of type "note") When tick = true, string with a percent sign indicating the position of the tick. Example: use tick_pos = "50%" for centered alignment.
- `tick_edge` - (Optional, only for widgets of type "note") When tick = true, string indicating on which side of the widget the tick should be displayed. One of "bottom", "top", "left", "right".
- `html` - (Optional, only for widgets of type "note") The content of the widget. HTML tags supported.
- `bgcolor` - (Optional, only for widgets of type "note") The color of the background of the widget.
- `event_size` - (Optional, only for widgets of type "event_stream") The size of the events in the widget. Either "s" (small, title only) or "l" (large, full event).
- `sizing` - (Optional, only for widgets of type "image") The preferred method to adapt the dimensions of the image to those of the widget. One of "center" (center the image in the tile), "zoom" (zoom the image to cover the whole tile) or "fit" (fit the image dimensions to those of the tile).
- `margin` - (Optional, only for widgets of type "image") The margins to use around the image. Either "small" or "large".
- `env` - (Optional, only for widgets of type "trace_service") The environment to use.
- `service_service` - (Optional, only for widgets of type "trace_service") The trace service to use.
- `service_name` - (Optional, only for widgets of type "trace_service") The name of the service to use.
- `size_version` - (Optional, only for widgets of type "trace_service") The size of the widget. One of "small", "medium", "large".
- `layout_version` - (Optional, only for widgets of type "trace_service") The number of columns to use when displaying values. One of "one_column", "two_column", "three_column".
- `must_show_hits` - (Optional, only for widgets of type "trace_service") Boolean indicating whether to display hits.
- `must_show_errors` - (Optional, only for widgets of type "trace_service") Boolean indicating whether to display errors.
- `must_show_latency` - (Optional, only for widgets of type "trace_service") Boolean indicating whether to display latency.
- `must_show_breakdown` - (Optional, only for widgets of type "trace_service") Boolean indicating whether to display breakdown.
- `must_show_distribution` - (Optional, only for widgets of type "trace_service") Boolean indicating whether to display distribution.
- `must_show_resource_list` - (Optional, only for widgets of type "trace_service") Boolean indicating whether to display resources.
- `display_format` - (Optional, only for widgets of type "manage_status") The display setting to use. One of "counts", "list", or "countsAndList".
- `color_preference` - (Optional, only for widgets of type "manage_status") Whether to colorize text or background. One of "text", "background".
- `hide_zero_counts` - (Optional, only for widgets of type "manage_status") Boolean indicating whether to hide empty categories.
- `manage_status_show_title` - (Optional, only for widgets of type "manage_status") Boolean indicating whether to show a title.
- `manage_status_title_text` - (Optional, only for widgets of type "manage_status") The title of the widget.
- `manage_status_title_size` - (Optional, only for widgets of type "manage_status") The size of the widget's title.
- `manage_status_title_align` - (Optional, only for widgets of type "manage_status") The alignment of the widget's title. One of "left", "center", or "right".
- `columns` - (Optional, only for widgets of type "log_stream") Stringified list of columns to use. Example: `"[\"column1\",\"column2\",\"column3\"]"`
- `logset` - (Optional, only for widgets of type "log_stream") ID of the logset to use.
- `time` - (Optional, only for widgets of type "timeseries", "toplist", "event_timeline", "event_stream", "alert_graph", "check_status", "trace_service", "log_stream") Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. At most one such block should be present in a given widget.
- `tile_def` - (Optional, only for widgets of type "timeseries", "query_value", "hostmap", "change", "toplist", "process") Nested block describing the content to display in the widget. The structure of this block is described below. At most one such block should be present in a given widget.
- `params` - (Optional, only for widgets of type "manage_status") Nested block describing the monitors to display. The structure of this block is described below. At most one such block should be present in a given widget.

### Nested `widget` `time` blocks

Only for widgets of type "timeseries", "toplist", "event_timeline", "event_stream", "alert_graph", "check_status", "trace_service", "log_stream".

Nested `widget` `time` blocks have the following structure:

- `live_span` - (Required) The timeframe to use when displaying the widget. One of "10m", "30m", "1h", "4h", "1d", "2d", "1w".

### Nested `widget` `params` blocks

Only for widgets of type "manage_status".

Nested `widget` `params` blocks have the following structure:

- `sort` - (Optional) The method to use to sort monitors. Example: "status,asc".
- `text` - (Optional) The query to use to get monitors. Example: "status:alert".
- `count` - (Optional) The number of monitors to display.
- `start` - (Optional) The start of the list. Typically 0.

### Nested `widget` `tile_def` blocks

Only for widgets of type "timeseries", "query_value", "hostmap", "change", "toplist", "process".

Nested `widget` `tile_def` blocks have the following structure:

- `viz` - (Required) Should be the same as the widget's type. One of "timeseries", "query_value", "hostmap", "change", "toplist", "process".
- `request` - (Required) Nested block describing the request to use when displaying the widget. The structure of this block is described below. Multiple request blocks are allowed within a given tile_def block.
- `marker` - (Optional, only for widgets of type "timeseries") Nested block describing the marker to use when displaying the widget. The structure of this block is described below. Multiple marker blocks are allowed within a given tile_def block.
- `event` - (Optional, only for widgets of type "timeseries") Nested block describing the event overlays to use when displaying the widget. The structure of this block is described below. At most one such block should be present in a given tile_def block.
- `custom_unit` - (Optional, only for widgets of type "query_value") The unit for the value displayed in the widget
- `autoscale` - (Optional, only for widgets of type "query_value") Boolean indicating whether to automatically scale the tile.
- `precision` - (Optional, only for widgets of type "query_value") The precision to use when displaying the tile.
- `text_align` - (Optional, only for widgets of type "query_value") The alignment of the text.
- `node_type` - (Optional, only for widgets of type "hostmap") The type of node used. Either "host" or "container".
- `scope` - (Optional, only for widgets of type "hostmap") The list of tags to filter nodes by.
- `group` - (Optional, only for widgets of type "hostmap") The list of tags to group nodes by.
- `no_group_host` - (Optional, only for widgets of type "hostmap") Boolean indicating whether to show ungrouped nodes.
- `no_metric_host` - (Optional, only for widgets of type "hostmap") Boolean indicating whether to show nodes with no metrics.
- `style` - (Optional, only for widgets of type "hostmap") Nested block describing how to display the widget. The structure of this block is described below. At most one such block should be present in a given tile_def block.

### Nested `widget` `tile_def` `style` blocks

Only for widgets of type "hostmap".

Nested `widget` `tile_def` `style` blocks have the following structure:

- `palette` - (Optional) Color set to use to display nodes. One of "green_to_orange", "yellow_to_green", "YlOrRd" (warm), "hostmap_blues" (cool).
- `palette_flip` - (Optional) Boolean indicating whether to flip how the hostmap is rendered. For example, with the default palette, low values are represented as green, with high values as orange. If palette_flip is "true", then low values will be orange, and high values will be green.
- `fill_min` - (Optional) Metric value corresponding to minimum color fill.
- `fill_max` - (Optional) Metric value corresponding to maximum color fill.

### Nested `widget` `tile_def` `marker` blocks

Only for widgets of type "timeseries".

Nested `widget` `tile_def` `marker` blocks have the following structure:

- `type` - (Required) How the marker lines will look. Possible values are {"error", "warning", "info", "ok"} {"dashed", "solid", "bold"}. Example: "error dashed".
- `value` - (Required) Mathematical expression describing the marker. Examples: "y > 1", "-5 < y < 0", "y = 19".
- `label` - (Optional) A label for the line or range.

### Nested `widget` `tile_def` `event` block

Only for widgets of type "timeseries".

Nested `widget` `tile_def` `event` blocks have the following structure:

- `q` - (Required) The search query for event overlays.

### Nested `widget` `tile_def` `request` blocks

Only for widgets of type "timeseries", "query_value", "toplist", "change", "hostmap", "process".

Nested `widget` `tile_def` `request` blocks have the following structure:

- `q` - (Optional, only for widgets of type "timeseries", "query_value", "toplist", "change", "hostmap") The query of the request. Pro tip: Use the JSON tab inside the Datadog UI to help build you query strings.
- `type` - (Optional, only for widgets of type "timeseries", "query_value", "hostmap") Choose the type of representation to use for this query. For widgets of type "timeseries" and "query_value", use one of "line", "bars" or "area". For widgets of type "hostmap", use "fill" or "size".
- `query_type` - (Optional, only for widgets of type "process") Use "process".
- `metric` - (Optional, only for widgets of type "process") The metric you want to use for the widget.
- `text_filter` - (Optional, only for widgets of type "process") The search query for the widget.
- `tag_filters` - (Optional, only for widgets of type "process") Tags to use for filtering.
- `limit` - (Optional, only for widgets of type "process") Integer indicating the number of hosts to limit to.
- `aggregator` - (Optional, only for widgets of type "query_value") The aggregator to use for time aggregation. One of "avg", "min", "max", "sum", "last".
- `compare_to` - (Optional, only for widgets of type "change") Choose from when to compare current data to. One of "hour_before", "day_before", "week_before" or "month_before".
- `change_type` - (Optional, only for widgets of type "change") Whether to show absolute or relative change. One of "absolute", "relative".
- `order_by` - (Optional, only for widgets of type "change") One of "change", "name", "present" (present value) or "past" (past value).
- `order_dir` - (Optional, only for widgets of type "change") Either "asc" (ascending) or "desc" (descending).
- `extra_col` - (Optional, only for widgets of type "change") If set to "present", displays current value. Can be left empty otherwise.
- `increase_good` - (Optional, only for widgets of type "change") Boolean indicating whether an increase in the value is good (thus displayed in green) or not (thus displayed in red).
- `style` - (Optional, only for widgets of type "timeseries", "query_value", "toplist", "process") describing how to display the widget. The structure of this block is described below. At most one such block should be present in a given request block.
- `conditional_format` - (Optional) Nested block to customize the style if certain conditions are met. Currently only applies to `Query Value` and `Top List` type graphs.
* `metadata_json` - (Optional) A JSON blob (preferrably created using [jsonencode](https://www.terraform.io/docs/configuration/functions/jsonencode.html)) representing mapping of query expressions to alias names. Note that the query expressions in `metadata_json` will be ignored if they're not present in the query. For example, this is how you define `metadata_json` with Terraform >= 0.12:
  ```
  metadata_json = jsonencode({
    "avg:redis.info.latency_ms{$host}": {
      "alias": "Redis latency"
    }
  })
  ```
  And here's how you define `metadata_json` with Terraform < 0.12:
  ```
  variable "my_metadata" {
    default = {
      "avg:redis.info.latency_ms{$host}" = {
        "alias": "Redis latency"
      }
    }
  }

  resource "datadog_screenboard" "SomeScreenboard" {
    ...
          metadata_json = "${jsonencode(var.my_metadata)}"
  }
  ```
  Note that this has to be a JSON blob because of [limitations](https://github.com/hashicorp/terraform/issues/6215) of Terraform's handling complex nested structures. This is also why the key is called `metadata_json` even though it sets `metadata` attribute on the API call.

### Nested `widget` `tile_def` `request` `style` block

Only for widgets of type "timeseries", "query_value", "toplist", "process".

The nested `style` blocks has the following structure:

- `palette` - (Optional) Color of the line drawn. For widgets of type "timeseries", "query_value", "toplist", one of: "classic", "cool", "warm", "purple", "orange" or "gray". For widgets of type "process", one of: "dog_classic_area", "YlOrRd", "GnBu", "Reds", "Oranges", "Greens", "Blues", "Purples".
- `width` - (Optional) Line width. Possible values: "thin", "normal", "thick". Default: "normal".
- `type` - (Optional) Type of line drawn. Possible values: "dashed", "solid", "dotted". Default: "solid".

### Nested `widget` `tile_def` `request` `conditional_format` block

The nested `conditional_format` blocks has the following structure:

- `palette` - (Optional) Color scheme to be used if the condition is met. One of: "red_on_white", "white_on_red", "yellow_on_white", "white_on_yellow", "green_on_white", "white_on_green", "gray_on_white", "white_on_gray", "custom_text", "custom_bg", "custom_image".
- `comparator` - (Required) Comparison operator. Example: ">", "<".
- `value` - (Optional) Value that is the threshold for the conditional format.
- `color` - (Optional) Custom color (e.g., #205081).
- `invert` - (Optional) Boolean indicating whether to invert color scheme.

### Nested `template_variable` blocks

Nested `template_variable` blocks have the following structure:

- `name` - (Required) The variable name. Can be referenced as $name in `graph` `request` `q` query strings.
- `prefix` - (Optional) The tag group. Default: no tag group.
- `default` - (Optional) The default tag. Default: "\*" (match all).

## Attributes Reference

The following attributes are exported:

- `id` - The unique ID of this screenboard in your Datadog account. The web interface URL to this screenboard can be generated by appending this ID to `https://app.datadoghq.com/screen/`

## Import

screenboards can be imported using their numeric ID, e.g.

```
$ terraform import datadog_screenboard.my_service_screenboard 2081
```
