---
page_title: "datadog_screenboard Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog screenboard resource. This can be used to create and manage Datadog screenboards.
---

# Resource `datadog_screenboard`

Provides a Datadog screenboard resource. This can be used to create and manage Datadog screenboards.

## Example Usage

```terraform
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

      request {
        log_query {
          index = "mcnulty"
          compute {
            aggregation = "avg"
            facet = "@duration"
            interval = 5000
          }
          search {
            query = "status:info"
          }
          group_by {
            facet = "host"
            limit = 10
            sort {
              aggregation = "avg"
              order = "desc"
              facet = "@duration"
            }
          }
        }
        type = "area"
      }

      request {
        apm_query {
          index = "apm-search"
          compute {
            aggregation = "avg"
            facet = "@duration"
            interval = 5000
          }
          search {
            query = "type:web"
          }
          group_by {
            facet = "resource_name"
            limit = 50
            sort {
              aggregation = "avg"
              order = "desc"
              facet = "@string_query.interval"
            }
          }
        }
        type = "bars"
      }

      request {
        process_query {
          metric = "process.stat.cpu.total_pct"
          search_by = "error"
          filter_by = ["active"]
          limit = 50
        }
        type = "area"
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
    summary_type              = "monitors"
    display_format            = "countsAndList"
    color_preference          = "background"
    hide_zero_counts          = true
    show_last_triggered       = false
    manage_status_show_title  = false
    manage_status_title_text  = "test title"
    manage_status_title_size  = "20"
    manage_status_title_align = "right"

    params = {
      sort  = "status,asc"
      text  = "status:alert"
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

## Schema

### Required

- **title** (String) Name of the screenboard
- **widget** (Block List, Min: 1) A list of widget definitions. (see [below for nested schema](#nestedblock--widget))

### Optional

- **height** (String) Height of the screenboard
- **id** (String) The ID of this resource.
- **read_only** (Boolean) The read-only status of the screenboard. Default is `false`.
- **shared** (Boolean) Whether the screenboard is shared or not
- **template_variable** (Block List) A list of template variables for using Dashboard templating. (see [below for nested schema](#nestedblock--template_variable))
- **width** (String) Width of the screenboard

<a id="nestedblock--widget"></a>
### Nested Schema for `widget`

Required:

- **type** (String) The type of the widget. One of [ 'free_text', 'timeseries', 'query_value', 'toplist', 'change', 'event_timeline', 'event_stream', 'image', 'note', 'alert_graph', 'alert_value', 'iframe', 'check_status', 'trace_service', 'hostmap', 'manage_status', 'log_stream', 'uptime', 'process']
- **x** (Number) The position of the widget on the x axis.
- **y** (Number) The position of the widget on the y axis.

Optional:

- **alert_id** (Number) Only for widgets of type `alert_value`, `alert_graph`. The ID of the monitor used by the widget.
- **auto_refresh** (Boolean) Only for widgets of type `alert_value`, `alert_graph`. Boolean indicating whether the widget is refreshed automatically.
- **bgcolor** (String) Only for widgets of type `note`. The color of the background of the widget.
- **check** (String) Only for widgets of type `check_status`. The check to use in the widget.
- **color** (String) Only for widgets of type `free_text`. The color of the text in the widget.
- **color_preference** (String) One of: ['background', 'text']
- **columns** (String) Only for widgets of type `log_stream`. Stringified list of columns to use. Example: `["column1","column2","column3"]`.
- **display_format** (String) One of: ['counts', 'list', 'countsAndList']
- **env** (String) Only for widgets of type `trace_service`. The environment to use.
- **event_size** (String) Only for widgets of type `event_stream`. The size of the events in the widget. Either `s` (small, title only) or `l` (large, full event).
- **font_size** (String) Only for widgets of type `free_text`, `note`. The size of the text in the widget.
- **group** (String) Only for widgets of type `check_status`. The check group to use in the widget.
- **group_by** (List of String) Only for widgets of type `check_status`. When `grouping = "cluster"`, indicates a list of tags to use for grouping.
- **grouping** (String) One of: ['cluster', 'check']
- **height** (Number) The height of the widget.
- **hide_zero_counts** (Boolean) Only for widgets of type `manage_status`. Boolean indicating whether to hide empty categories.
- **html** (String) Only for widgets of type `note`. The content of the widget. HTML tags supported.
- **layout_version** (String) Only for widgets of type `trace_service`. The number of columns to use when displaying values. One of `one_column`, `two_column`, `three_column`.
- **legend** (Boolean) Only for widgets of type `timeseries`, `query_value`, `toplist`. Boolean indicating whether to display a legend in the widget.
- **legend_size** (String) Only for widgets of type `timeseries`, `query_value`, `toplist`. The size of the legend displayed in the widget.
- **logset** (String) Only for widgets of type `log_stream`. ID of the logset to use.
- **manage_status_show_title** (Boolean) Only for widgets of type `manage_status`. Boolean indicating whether to show a title.
- **manage_status_title_align** (String) Only for widgets of type `manage_status`. The alignment of the widget's title. One of `left`, `center`, or `right`.
- **manage_status_title_size** (String) Only for widgets of type `manage_status`. The size of the widget's title.
- **manage_status_title_text** (String) Only for widgets of type `manage_status`. The title of the widget.
- **margin** (String) One of: ['small', 'large']
- **monitor** (Map of String)
- **must_show_breakdown** (Boolean) Only for widgets of type `trace_service`. Boolean indicating whether to display breakdown.
- **must_show_distribution** (Boolean) Only for widgets of type `trace_service`. Boolean indicating whether to display distribution.
- **must_show_errors** (Boolean) Only for widgets of type `trace_service`. Boolean indicating whether to display errors.
- **must_show_hits** (Boolean) Only for widgets of type `trace_service`. Boolean indicating whether to display hits.
- **must_show_latency** (Boolean) Only for widgets of type `trace_service`. Boolean indicating whether to display latency.
- **must_show_resource_list** (Boolean) Only for widgets of type `trace_service` Boolean indicating whether to display resources.
- **params** (Map of String) Only for widgets of type `manage_status`. Nested block describing the monitors to display. The structure of this block is described below. At most one such block should be present in a given widget.
- **precision** (String) Only for widgets of type `query_value`. The precision to use when displaying the tile.
- **query** (String) Only for widgets of type `event_timeline`, `event_stream`, `hostmap`, `log_stream`. The query to use in the widget.
- **rule** (Block List) (see [below for nested schema](#nestedblock--widget--rule))
- **service_name** (String) Only for widgets of type `trace_service`. The name of the service to use.
- **service_service** (String) Only for widgets of type `trace_service`. The trace service to use.
- **show_last_triggered** (Boolean) Only for widgets of type `manage_status`. Boolean indicating whether to show when monitors/groups last triggered.
- **size_version** (String) Only for widgets of type `trace_service`. The size of the widget. One of `small`, `medium`, `large`.
- **sizing** (String) One of: ['center', 'zoom', 'fit']
- **summary_type** (String) One of: ['monitors', 'groups', 'combined']
- **tags** (List of String) Only for widgets of type `check_status`. List of tags to use in the widget.
- **text** (String) For widgets of type 'free_text', the text to use.
- **text_align** (String) Only for widgets of type `free_text`, `alert_value`, `note`. The alignment of the text in the widget.
- **text_size** (String) Only for widgets of type `alert_value`. The size of the text in the widget.
- **tick** (Boolean) Only for widgets of type `note`. Boolean indicating whether a tick should be displayed on the border of the widget.
- **tick_edge** (String) Only for widgets of type `note`. When `tick = true`, string indicating on which side of the widget the tick should be displayed. One of `bottom`, `top`, `left`, `right`.
- **tick_pos** (String) Only for widgets of type `note`. When `tick = true`, string with a percent sign indicating the position of the tick. Example: use `tick_pos = "50%"` for centered alignment.
- **tile_def** (Block List) (see [below for nested schema](#nestedblock--widget--tile_def))
- **time** (Map of String) Only for widgets of type `timeseries`, `toplist`, `event_timeline`, `event_stream`, `alert_graph`, `check_status`, `trace_service`, `log_stream`. Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. At most one such block should be present in a given widget.
- **timeframes** (List of String)
- **title** (String) The name of the widget.
- **title_align** (String) The alignment of the widget's title.
- **title_size** (Number) The size of the widget's title.
- **unit** (String) Only for widgets of type `alert_value`. The unit for the value displayed in the widget.
- **url** (String) Only for widgets of type `image`, `iframe`. The URL to use as a data source for the widget.
- **viz_type** (String) One of: ['timeseries', 'toplist']
- **width** (Number) The width of the widget.

<a id="nestedblock--widget--rule"></a>
### Nested Schema for `widget.rule`

Optional:

- **color** (String)
- **threshold** (Number)
- **timeframe** (String)


<a id="nestedblock--widget--tile_def"></a>
### Nested Schema for `widget.tile_def`

Required:

- **request** (Block List, Min: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request))
- **viz** (String) Should be the same as the widget's type. One of `timeseries`, `query_value`, `hostmap`, `change`, `toplist`, `process`.

Optional:

- **autoscale** (Boolean) Only for widgets of type `query_value`. Boolean indicating whether to automatically scale the tile.
- **custom_unit** (String) Only for widgets of type `query_value`. The unit for the value displayed in the widget.
- **event** (Block List) (see [below for nested schema](#nestedblock--widget--tile_def--event))
- **group** (List of String) Only for widgets of type `hostmap`. The list of tags to group nodes by.
- **marker** (Block List) (see [below for nested schema](#nestedblock--widget--tile_def--marker))
- **no_group_hosts** (Boolean) Only for widgets of type `hostmap`. Boolean indicating whether to show ungrouped nodes.
- **no_metric_hosts** (Boolean) Only for widgets of type `hostmap`. Boolean indicating whether to show nodes with no metrics.
- **node_type** (String) Only for widgets of type `hostmap`. The type of node used. Either `host` or `container`.
- **precision** (String) Only for widgets of type `query_value`. The precision to use when displaying the tile.
- **scope** (List of String) Only for widgets of type `hostmap`. The list of tags to filter nodes by.
- **style** (Map of String) Only for widgets of type `hostmap`. Nested block describing how to display the widget. The structure of this block is described below. At most one such block should be present in a given `tile_def` block.
- **text_align** (String) Only for widgets of type `query_value`. The alignment of the text.

<a id="nestedblock--widget--tile_def--request"></a>
### Nested Schema for `widget.tile_def.request`

Optional:

- **aggregator** (String) Only for widgets of type `query_value`. The aggregator to use for time aggregation. One of `avg`, `min`, `max`, `sum`, `last`.
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--apm_query))
- **change_type** (String) Only for widgets of type `change`. Whether to show absolute or relative change. One of `absolute`, `relative`.
- **compare_to** (String) Only for widgets of type `change`. Choose from when to compare current data to. One of `hour_before`, `day_before`, `week_before` or `month_before`.
- **conditional_format** (Block List) A list of conditional formatting rules. (see [below for nested schema](#nestedblock--widget--tile_def--request--conditional_format))
- **extra_col** (String) Only for widgets of type `change`. If set to `present`, displays current value. Can be left empty otherwise.
- **increase_good** (Boolean) Only for widgets of type `change`. Boolean indicating whether an increase in the value is good (thus displayed in green) or not (thus displayed in red).
- **limit** (Number) Only for widgets of type `process`. Integer indicating the number of hosts to limit to.
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--log_query))
- **metadata_json** (String) A JSON blob (preferrably created using [jsonencode](https://www.terraform.io/docs/configuration/functions/jsonencode.html?_ga=2.6381362.1091155358.1609189257-888022054.1605547463)) representing mapping of query expressions to alias names. Note that the query expressions in `metadata_json` will be ignored if they're not present in the query.
- **metric** (String) Only for widgets of type `process`. The metric you want to use for the widget.
- **order_by** (String) Only for widgets of type `change`. One of `change`, `name`, `present` (present value) or `past` (past value).
- **order_dir** (String) Only for widgets of type `change`. Either `asc` (ascending) or `desc` (descending).
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--process_query))
- **q** (String) Only for widgets of type `timeseries`, `query_value`, `toplist`, `change`, `hostmap`: The query of the request. Pro tip: Use the JSON tab inside the Datadog UI to help build you query strings.
- **query_type** (String) Only for widgets of type `process`. Use `process`.
- **style** (Map of String) Only for widgets of type `timeseries`, `query_value`, `toplist`, `process`. How to display the widget. The structure of this block is described below. At most one such block should be present in a given `request` block.
- **tag_filters** (List of String) Only for widgets of type `process`. Tags to use for filtering.
- **text_filter** (String) Only for widgets of type `process`. The search query for the widget.
- **type** (String) Only for widgets of type `timeseries`, `query_value`, `hostmap`: Choose the type of representation to use for this query. For widgets of type `timeseries` and `query_value`, use one of `line`, `bars` or `area`. For widgets of type `hostmap`, use `fill` or `size`.

<a id="nestedblock--widget--tile_def--request--apm_query"></a>
### Nested Schema for `widget.tile_def.request.type`

Required:

- **compute** (Block List, Min: 1, Max: 1) Exactly one nested block is required with the structure below. (see [below for nested schema](#nestedblock--widget--tile_def--request--type--compute))
- **index** (String) Name of the index to query

Optional:

- **group_by** (Block List) Multiple nested blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--tile_def--request--type--group_by))
- **search** (Block List, Max: 1) One nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--tile_def--request--type--search))

<a id="nestedblock--widget--tile_def--request--type--compute"></a>
### Nested Schema for `widget.tile_def.request.type.compute`

Required:

- **aggregation** (String) The aggregation method.

Optional:

- **facet** (String) Facet name.
- **interval** (String) Define a time interval in seconds.


<a id="nestedblock--widget--tile_def--request--type--group_by"></a>
### Nested Schema for `widget.tile_def.request.type.group_by`

Required:

- **facet** (String) Facet name.

Optional:

- **limit** (Number) Maximum number of items in the group.
- **sort** (Block List, Max: 1) One map is allowed with the keys as below. (see [below for nested schema](#nestedblock--widget--tile_def--request--type--group_by--sort))

<a id="nestedblock--widget--tile_def--request--type--group_by--sort"></a>
### Nested Schema for `widget.tile_def.request.type.group_by.sort`

Required:

- **aggregation** (String) The aggregation method.
- **order** (String) Widget sorting methods.

Optional:

- **facet** (String) Facet name.



<a id="nestedblock--widget--tile_def--request--type--search"></a>
### Nested Schema for `widget.tile_def.request.type.search`

Required:

- **query** (String) Query to use.



<a id="nestedblock--widget--tile_def--request--conditional_format"></a>
### Nested Schema for `widget.tile_def.request.type`

Required:

- **comparator** (String) Comparator (<, >, etc)

Optional:

- **color** (String) Custom color (e.g., #205081)
- **custom_bg_color** (String) Custom  background color (e.g., #205081)
- **invert** (Boolean) Boolean indicating whether to invert color scheme.
- **palette** (String) The palette to use if this condition is met.
- **value** (String) Value that is threshold for conditional format


<a id="nestedblock--widget--tile_def--request--log_query"></a>
### Nested Schema for `widget.tile_def.request.type`

Required:

- **compute** (Block List, Min: 1, Max: 1) Exactly one nested block is required with the structure below. (see [below for nested schema](#nestedblock--widget--tile_def--request--type--compute))
- **index** (String) Name of the index to query

Optional:

- **group_by** (Block List) Multiple nested blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--tile_def--request--type--group_by))
- **search** (Block List, Max: 1) One nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--tile_def--request--type--search))

<a id="nestedblock--widget--tile_def--request--type--compute"></a>
### Nested Schema for `widget.tile_def.request.type.compute`

Required:

- **aggregation** (String) The aggregation method.

Optional:

- **facet** (String) Facet name.
- **interval** (String) Define a time interval in seconds.


<a id="nestedblock--widget--tile_def--request--type--group_by"></a>
### Nested Schema for `widget.tile_def.request.type.group_by`

Required:

- **facet** (String) Facet name.

Optional:

- **limit** (Number) Maximum number of items in the group.
- **sort** (Block List, Max: 1) One map is allowed with the keys as below. (see [below for nested schema](#nestedblock--widget--tile_def--request--type--group_by--sort))

<a id="nestedblock--widget--tile_def--request--type--group_by--sort"></a>
### Nested Schema for `widget.tile_def.request.type.group_by.sort`

Required:

- **aggregation** (String) The aggregation method.
- **order** (String) Widget sorting methods.

Optional:

- **facet** (String) Facet name.



<a id="nestedblock--widget--tile_def--request--type--search"></a>
### Nested Schema for `widget.tile_def.request.type.search`

Required:

- **query** (String) Query to use.



<a id="nestedblock--widget--tile_def--request--process_query"></a>
### Nested Schema for `widget.tile_def.request.type`

Required:

- **metric** (String) Your chosen metric.

Optional:

- **filter_by** (List of String) List of processes.
- **limit** (Number) Max number of items in the filter list.
- **search_by** (String) Your chosen search term.



<a id="nestedblock--widget--tile_def--event"></a>
### Nested Schema for `widget.tile_def.event`

Required:

- **q** (String) The search query for event overlays.


<a id="nestedblock--widget--tile_def--marker"></a>
### Nested Schema for `widget.tile_def.marker`

Required:

- **type** (String) How the marker lines will look. Possible values are one of {`error`, `warning`, `info`, `ok`} combined with one of {`dashed`, `solid`, `bold`}. Example: `error dashed`.
- **value** (String) Mathematical expression describing the marker. Examples: `y > 1`, `-5 < y < 0`, `y = 19`.

Optional:

- **label** (String) A label for the line or range.




<a id="nestedblock--template_variable"></a>
### Nested Schema for `template_variable`

Required:

- **name** (String) The name of the variable.

Optional:

- **default** (String) The default value for the template variable on dashboard load.
- **prefix** (String) The tag prefix associated with the variable. Only tags with this prefix will appear in the variable dropdown.

## Import

Import is supported using the following syntax:

```shell
# screenboards can be imported using their numeric ID, e.g.
terraform import datadog_screenboard.my_service_screenboard 2081
```
