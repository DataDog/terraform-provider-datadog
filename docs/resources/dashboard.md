---
page_title: "datadog_dashboard Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog dashboard resource. This can be used to create and manage Datadog dashboards.
---

# Resource `datadog_dashboard`

Provides a Datadog dashboard resource. This can be used to create and manage Datadog dashboards.

## Example Usage

```terraform
# Example Ordered Layout
resource "datadog_dashboard" "ordered_dashboard" {
  title        = "Ordered Layout Dashboard"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "ordered"
  is_read_only = true

  widget {
    alert_graph_definition {
      alert_id  = "895605"
      viz_type  = "timeseries"
      title     = "Widget Title"
      live_span = "1h"
    }
  }

  widget {
    alert_value_definition {
      alert_id   = "895605"
      precision  = 3
      unit       = "b"
      text_align = "center"
      title      = "Widget Title"
    }
  }

  widget {
    alert_value_definition {
      alert_id   = "895605"
      precision  = 3
      unit       = "b"
      text_align = "center"
      title      = "Widget Title"
    }
  }

  widget {
    change_definition {
      request {
        q             = "avg:system.load.1{env:staging} by {account}"
        change_type   = "absolute"
        compare_to    = "week_before"
        increase_good = true
        order_by      = "name"
        order_dir     = "desc"
        show_present  = true
      }
      title     = "Widget Title"
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
      title     = "Widget Title"
      live_span = "1h"
    }
  }

  widget {
    check_status_definition {
      check     = "aws.ecs.agent_connected"
      grouping  = "cluster"
      group_by  = ["account", "cluster"]
      tags      = ["account:demo", "cluster:awseb-ruthebdog-env-8-dn3m6u3gvk"]
      title     = "Widget Title"
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
        min          = 1
        max          = 2
        include_zero = true
        scale        = "sqrt"
      }
      title     = "Widget Title"
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
      node_type       = "container"
      group           = ["host", "region"]
      no_group_hosts  = true
      no_metric_hosts = true
      scope           = ["region:us-east-1", "aws_account:727006795293"]
      style {
        palette      = "yellow_to_green"
        palette_flip = true
        fill_min     = "10"
        fill_max     = "20"
      }
      title = "Widget Title"
    }
  }

  widget {
    note_definition {
      content          = "note text"
      background_color = "pink"
      font_size        = "14"
      text_align       = "center"
      show_tick        = true
      tick_edge        = "left"
      tick_pos         = "50%"
    }
  }

  widget {
    query_value_definition {
      request {
        q          = "avg:system.load.1{env:staging} by {account}"
        aggregator = "sum"
        conditional_formats {
          comparator = "<"
          value      = "2"
          palette    = "white_on_green"
        }
        conditional_formats {
          comparator = ">"
          value      = "2.2"
          palette    = "white_on_red"
        }
      }
      autoscale   = true
      custom_unit = "xx"
      precision   = "4"
      text_align  = "right"
      title       = "Widget Title"
      live_span   = "1h"
    }
  }

  widget {
    query_table_definition {
      request {
        q          = "avg:system.load.1{env:staging} by {account}"
        aggregator = "sum"
        limit      = "10"
        conditional_formats {
          comparator = "<"
          value      = "2"
          palette    = "white_on_green"
        }
        conditional_formats {
          comparator = ">"
          value      = "2.2"
          palette    = "white_on_red"
        }
      }
      title     = "Widget Title"
      live_span = "1h"
    }
  }

  widget {
    scatterplot_definition {
      request {
        x {
          q          = "avg:system.cpu.user{*} by {service, account}"
          aggregator = "max"
        }
        y {
          q          = "avg:system.mem.used{*} by {service, account}"
          aggregator = "min"
        }
      }
      color_by_groups = ["account", "apm-role-group"]
      xaxis {
        include_zero = true
        label        = "x"
        min          = "1"
        max          = "2000"
        scale        = "pow"
      }
      yaxis {
        include_zero = false
        label        = "y"
        min          = "5"
        max          = "2222"
        scale        = "log"
      }
      title     = "Widget Title"
      live_span = "1h"
    }
  }

  widget {
    servicemap_definition {
      service     = "master-db"
      filters     = ["env:prod", "datacenter:us1.prod.dog"]
      title       = "env: prod, datacenter:us1.prod.dog, service: master-db"
      title_size  = "16"
      title_align = "left"
    }
    widget_layout {
      height = 43
      width  = 32
      x      = 5
      y      = 5
    }
  }

  widget {
    timeseries_definition {
      request {
        q            = "avg:system.cpu.user{app:general} by {env}"
        display_type = "line"
        style {
          palette    = "warm"
          line_type  = "dashed"
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
            aggregation = "avg"
            facet       = "@duration"
            interval    = 5000
          }
          search_query = "status:info"
          group_by {
            facet = "host"
            limit = 10
            sort_query {
              aggregation = "avg"
              order       = "desc"
              facet       = "@duration"
            }
          }
        }
        display_type = "area"
      }
      request {
        apm_query {
          index = "apm-search"
          compute_query {
            aggregation = "avg"
            facet       = "@duration"
            interval    = 5000
          }
          search_query = "type:web"
          group_by {
            facet = "resource_name"
            limit = 50
            sort_query {
              aggregation = "avg"
              order       = "desc"
              facet       = "@string_query.interval"
            }
          }
        }
        display_type = "bars"
      }
      request {
        process_query {
          metric    = "process.stat.cpu.total_pct"
          search_by = "error"
          filter_by = ["active"]
          limit     = 50
        }
        display_type = "area"
      }
      marker {
        display_type = "error dashed"
        label        = " z=6 "
        value        = "y = 4"
      }
      marker {
        display_type = "ok solid"
        value        = "10 < y < 999"
        label        = " x=8 "
      }
      title       = "Widget Title"
      show_legend = true
      legend_size = "2"
      live_span   = "1h"
      event {
        q = "sources:test tags:1"
      }
      event {
        q = "sources:test tags:2"
      }
      yaxis {
        scale        = "log"
        include_zero = false
        max          = 100
      }
    }
  }

  widget {
    toplist_definition {
      request {
        q = "avg:system.cpu.user{app:general} by {env}"
        conditional_formats {
          comparator = "<"
          value      = "2"
          palette    = "white_on_green"
        }
        conditional_formats {
          comparator = ">"
          value      = "2.2"
          palette    = "white_on_red"
        }
      }
      title = "Widget Title"
    }
  }

  widget {
    group_definition {
      layout_type = "ordered"
      title       = "Group Widget"

      widget {
        note_definition {
          content          = "cluster note widget"
          background_color = "pink"
          font_size        = "14"
          text_align       = "center"
          show_tick        = true
          tick_edge        = "left"
          tick_pos         = "50%"
        }
      }

      widget {
        alert_graph_definition {
          alert_id  = "123"
          viz_type  = "toplist"
          title     = "Alert Graph"
          live_span = "1h"
        }
      }
    }
  }

  widget {
    service_level_objective_definition {
      title             = "Widget Title"
      view_type         = "detail"
      slo_id            = "56789"
      show_error_budget = true
      view_mode         = "overall"
      time_windows      = ["7d", "previous_week"]
    }
  }

  template_variable {
    name    = "var_1"
    prefix  = "host"
    default = "aws"
  }
  template_variable {
    name    = "var_2"
    prefix  = "service_name"
    default = "autoscaling"
  }

  template_variable_preset {
    name = "preset_1"
    template_variable {
      name  = "var_1"
      value = "host.dc"
    }
    template_variable {
      name  = "var_2"
      value = "my_service"
    }
  }
}

# Example Free Layout
resource "datadog_dashboard" "free_dashboard" {
  title        = "Free Layout Dashboard"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "free"
  is_read_only = false

  widget {
    event_stream_definition {
      query       = "*"
      event_size  = "l"
      title       = "Widget Title"
      title_size  = 16
      title_align = "left"
      live_span   = "1h"
    }
    widget_layout {
      height = 43
      width  = 32
      x      = 5
      y      = 5
    }
  }

  widget {
    event_timeline_definition {
      query       = "*"
      title       = "Widget Title"
      title_size  = 16
      title_align = "left"
      live_span   = "1h"
    }
    widget_layout {
      height = 9
      width  = 65
      x      = 42
      y      = 73
    }
  }

  widget {
    free_text_definition {
      text       = "free text content"
      color      = "#d00"
      font_size  = "88"
      text_align = "left"
    }
    widget_layout {
      height = 20
      width  = 30
      x      = 42
      y      = 5
    }
  }

  widget {
    iframe_definition {
      url = "http://google.com"
    }
    widget_layout {
      height = 46
      width  = 39
      x      = 111
      y      = 8
    }
  }

  widget {
    image_definition {
      url    = "https://images.pexels.com/photos/67636/rose-blue-flower-rose-blooms-67636.jpeg?auto=compress&cs=tinysrgb&h=350"
      sizing = "fit"
      margin = "small"
    }
    widget_layout {
      height = 20
      width  = 30
      x      = 77
      y      = 7
    }
  }

  widget {
    log_stream_definition {
      indexes             = ["main"]
      query               = "error"
      columns             = ["core_host", "core_service", "tag_source"]
      show_date_column    = true
      show_message_column = true
      message_display     = "expanded-md"
      sort {
        column = "time"
        order  = "desc"
      }
    }
    widget_layout {
      height = 36
      width  = 32
      x      = 5
      y      = 51
    }
  }

  widget {
    manage_status_definition {
      color_preference    = "text"
      display_format      = "countsAndList"
      hide_zero_counts    = true
      query               = "type:metric"
      show_last_triggered = false
      sort                = "status,asc"
      summary_type        = "monitors"
      title               = "Widget Title"
      title_size          = 16
      title_align         = "left"
    }
    widget_layout {
      height = 40
      width  = 30
      x      = 112
      y      = 55
    }
  }

  widget {
    trace_service_definition {
      display_format     = "three_column"
      env                = "datad0g.com"
      service            = "alerting-cassandra"
      show_breakdown     = true
      show_distribution  = true
      show_errors        = true
      show_hits          = true
      show_latency       = false
      show_resource_list = false
      size_format        = "large"
      span_name          = "cassandra.query"
      title              = "alerting-cassandra #env:datad0g.com"
      title_align        = "center"
      title_size         = "13"
      live_span          = "1h"
    }
    widget_layout {
      height = 38
      width  = 67
      x      = 40
      y      = 28
    }
  }

  template_variable {
    name    = "var_1"
    prefix  = "host"
    default = "aws"
  }
  template_variable {
    name    = "var_2"
    prefix  = "service_name"
    default = "autoscaling"
  }

  template_variable_preset {
    name = "preset_1"
    template_variable {
      name  = "var_1"
      value = "host.dc"
    }
    template_variable {
      name  = "var_2"
      value = "my_service"
    }
  }
}
```

## Schema

### Required

- **layout_type** (String, Required) The layout type of the dashboard, either 'free' or 'ordered'.
- **title** (String, Required) The title of the dashboard.
- **widget** (Block List, Min: 1) The list of widgets to display on the dashboard. (see [below for nested schema](#nestedblock--widget))

### Optional

- **dashboard_lists** (Set of Number, Optional) The list of dashboard lists this dashboard belongs to.
- **description** (String, Optional) The description of the dashboard.
- **id** (String, Optional) The ID of this resource.
- **is_read_only** (Boolean, Optional) Whether this dashboard is read-only.
- **notify_list** (List of String, Optional) The list of handles of users to notify when changes are made to this dashboard.
- **template_variable** (Block List) The list of template variables for this dashboard. (see [below for nested schema](#nestedblock--template_variable))
- **template_variable_preset** (Block List) The list of selectable template variable presets for this dashboard. (see [below for nested schema](#nestedblock--template_variable_preset))
- **url** (String, Optional) The URL of the dashboard.

### Read-only

- **dashboard_lists_removed** (Set of Number, Read-only) The list of dashboard lists this dashboard should be removed from. Internal only.

<a id="nestedblock--widget"></a>
### Nested Schema for `widget`

Optional:

- **alert_graph_definition** (Block List, Max: 1) The definition for a Alert Graph widget. (see [below for nested schema](#nestedblock--widget--alert_graph_definition))
- **alert_value_definition** (Block List, Max: 1) The definition for a Alert Value widget. (see [below for nested schema](#nestedblock--widget--alert_value_definition))
- **change_definition** (Block List, Max: 1) The definition for a Change  widget. (see [below for nested schema](#nestedblock--widget--change_definition))
- **check_status_definition** (Block List, Max: 1) The definition for a Check Status widget. (see [below for nested schema](#nestedblock--widget--check_status_definition))
- **distribution_definition** (Block List, Max: 1) The definition for a Distribution widget. (see [below for nested schema](#nestedblock--widget--distribution_definition))
- **event_stream_definition** (Block List, Max: 1) The definition for a Event Stream widget. (see [below for nested schema](#nestedblock--widget--event_stream_definition))
- **event_timeline_definition** (Block List, Max: 1) The definition for a Event Timeline widget. (see [below for nested schema](#nestedblock--widget--event_timeline_definition))
- **free_text_definition** (Block List, Max: 1) The definition for a Free Text widget. (see [below for nested schema](#nestedblock--widget--free_text_definition))
- **group_definition** (Block List, Max: 1) The definition for a Group widget. (see [below for nested schema](#nestedblock--widget--group_definition))
- **heatmap_definition** (Block List, Max: 1) The definition for a Heatmap widget. (see [below for nested schema](#nestedblock--widget--heatmap_definition))
- **hostmap_definition** (Block List, Max: 1) The definition for a Hostmap widget. (see [below for nested schema](#nestedblock--widget--hostmap_definition))
- **iframe_definition** (Block List, Max: 1) The definition for an Iframe widget. (see [below for nested schema](#nestedblock--widget--iframe_definition))
- **image_definition** (Block List, Max: 1) The definition for an Image widget (see [below for nested schema](#nestedblock--widget--image_definition))
- **layout** (Map of String, Optional, Deprecated) The layout of the widget on a 'free' dashboard. **Deprecated.** Define `widget_layout` list with one element instead.
- **log_stream_definition** (Block List, Max: 1) The definition for an Log Stream widget. (see [below for nested schema](#nestedblock--widget--log_stream_definition))
- **manage_status_definition** (Block List, Max: 1) The definition for an Manage Status widget. (see [below for nested schema](#nestedblock--widget--manage_status_definition))
- **note_definition** (Block List, Max: 1) The definition for a Note widget. (see [below for nested schema](#nestedblock--widget--note_definition))
- **query_table_definition** (Block List, Max: 1) The definition for a Query Table widget. (see [below for nested schema](#nestedblock--widget--query_table_definition))
- **query_value_definition** (Block List, Max: 1) The definition for a Query Value widget. (see [below for nested schema](#nestedblock--widget--query_value_definition))
- **scatterplot_definition** (Block List, Max: 1) The definition for a Scatterplot widget. (see [below for nested schema](#nestedblock--widget--scatterplot_definition))
- **service_level_objective_definition** (Block List, Max: 1) The definition for a Service Level Objective widget. (see [below for nested schema](#nestedblock--widget--service_level_objective_definition))
- **servicemap_definition** (Block List, Max: 1) The definition for a Service Map widget. (see [below for nested schema](#nestedblock--widget--servicemap_definition))
- **timeseries_definition** (Block List, Max: 1) The definition for a Timeseries widget. (see [below for nested schema](#nestedblock--widget--timeseries_definition))
- **toplist_definition** (Block List, Max: 1) The definition for a Toplist widget. (see [below for nested schema](#nestedblock--widget--toplist_definition))
- **trace_service_definition** (Block List, Max: 1) The definition for a Trace Service widget. (see [below for nested schema](#nestedblock--widget--trace_service_definition))
- **widget_layout** (Block List, Max: 1) The layout of the widget on a 'free' dashboard. (see [below for nested schema](#nestedblock--widget--widget_layout))

Read-only:

- **id** (Number, Read-only) The ID of the widget.

<a id="nestedblock--widget--alert_graph_definition"></a>
### Nested Schema for `widget.alert_graph_definition`

Required:

- **alert_id** (String, Required) The ID of the monitor used by the widget.
- **viz_type** (String, Required) Type of visualization to use when displaying the widget. Either `timeseries` or `toplist`.

Optional:

- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--alert_value_definition"></a>
### Nested Schema for `widget.alert_value_definition`

Required:

- **alert_id** (String, Required) The ID of the monitor used by the widget.

Optional:

- **precision** (Number, Optional) The precision to use when displaying the value. Use `*` for maximum precision.
- **text_align** (String, Optional) The alignment of the text in the widget.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.
- **unit** (String, Optional) The unit for the value displayed in the widget.


<a id="nestedblock--widget--change_definition"></a>
### Nested Schema for `widget.change_definition`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--change_definition--custom_link))
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple request blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--change_definition--request))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--change_definition--custom_link"></a>
### Nested Schema for `widget.change_definition.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--change_definition--request"></a>
### Nested Schema for `widget.change_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--change_definition--request--apm_query))
- **change_type** (String, Optional) Whether to show absolute or relative change. One of `absolute`, `relative`.
- **compare_to** (String, Optional) Choose from when to compare current data to. One of `hour_before`, `day_before`, `week_before` or `month_before`.
- **increase_good** (Boolean, Optional) Boolean indicating whether an increase in the value is good (thus displayed in green) or not (thus displayed in red).
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--change_definition--request--log_query))
- **order_by** (String, Optional) One of `change`, `name`, `present` (present value) or `past` (past value).
- **order_dir** (String, Optional) Either `asc` (ascending) or `desc` (descending).
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--change_definition--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--change_definition--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--change_definition--request--security_query))
- **show_present** (Boolean, Optional) If set to `true`, displays current value.

<a id="nestedblock--widget--change_definition--request--apm_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--change_definition--request--show_present--compute_query"></a>
### Nested Schema for `widget.change_definition.request.show_present.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--change_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by--sort_query))

<a id="nestedblock--widget--change_definition--request--show_present--group_by--sort_query"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--change_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.change_definition.request.show_present.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--change_definition--request--log_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--change_definition--request--show_present--compute_query"></a>
### Nested Schema for `widget.change_definition.request.show_present.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--change_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by--sort_query))

<a id="nestedblock--widget--change_definition--request--show_present--group_by--sort_query"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--change_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.change_definition.request.show_present.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--change_definition--request--process_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--change_definition--request--rum_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--change_definition--request--show_present--compute_query"></a>
### Nested Schema for `widget.change_definition.request.show_present.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--change_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by--sort_query))

<a id="nestedblock--widget--change_definition--request--show_present--group_by--sort_query"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--change_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.change_definition.request.show_present.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--change_definition--request--security_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--change_definition--request--show_present--compute_query"></a>
### Nested Schema for `widget.change_definition.request.show_present.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--change_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by--sort_query))

<a id="nestedblock--widget--change_definition--request--show_present--group_by--sort_query"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--change_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.change_definition.request.show_present.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--check_status_definition"></a>
### Nested Schema for `widget.check_status_definition`

Required:

- **check** (String, Required) The check to use in the widget.
- **grouping** (String, Required) Either `check` or `cluster`, depending on whether the widget should use a single check or a cluster of checks.

Optional:

- **group** (String, Optional) The check group to use in the widget.
- **group_by** (List of String, Optional) When `grouping = "cluster"`, indicates a list of tags to use for grouping.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **tags** (List of String, Optional) List of tags to use in the widget.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--distribution_definition"></a>
### Nested Schema for `widget.distribution_definition`

Optional:

- **legend_size** (String, Optional) The size of the legend displayed in the widget.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple request blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--distribution_definition--request))
- **show_legend** (Boolean, Optional) Whether or not to show the legend on this widget.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--distribution_definition--request"></a>
### Nested Schema for `widget.distribution_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--security_query))
- **style** (Block List, Max: 1) Style of the widget graph. One nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style))

<a id="nestedblock--widget--distribution_definition--request--apm_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--distribution_definition--request--style--compute_query"></a>
### Nested Schema for `widget.distribution_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--distribution_definition--request--style--group_by"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--distribution_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--distribution_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.distribution_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--distribution_definition--request--log_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--distribution_definition--request--style--compute_query"></a>
### Nested Schema for `widget.distribution_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--distribution_definition--request--style--group_by"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--distribution_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--distribution_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.distribution_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--distribution_definition--request--process_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--distribution_definition--request--rum_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--distribution_definition--request--style--compute_query"></a>
### Nested Schema for `widget.distribution_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--distribution_definition--request--style--group_by"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--distribution_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--distribution_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.distribution_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--distribution_definition--request--security_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--distribution_definition--request--style--compute_query"></a>
### Nested Schema for `widget.distribution_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--distribution_definition--request--style--group_by"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--distribution_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--distribution_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.distribution_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--distribution_definition--request--style"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Optional:

- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.




<a id="nestedblock--widget--event_stream_definition"></a>
### Nested Schema for `widget.event_stream_definition`

Required:

- **query** (String, Required) The query to use in the widget.

Optional:

- **event_size** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **tags_execution** (String, Optional) The execution method for multi-value filters. Can be either `and` or `or`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--event_timeline_definition"></a>
### Nested Schema for `widget.event_timeline_definition`

Required:

- **query** (String, Required) The query to use in the widget.

Optional:

- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **tags_execution** (String, Optional) The execution method for multi-value filters. Can be either `and` or `or`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--free_text_definition"></a>
### Nested Schema for `widget.free_text_definition`

Required:

- **text** (String, Required) The text to display in the widget.

Optional:

- **color** (String, Optional) The color of the text in the widget.
- **font_size** (String, Optional) The size of the text in the widget.
- **text_align** (String, Optional) The alignment of the text in the widget.


<a id="nestedblock--widget--group_definition"></a>
### Nested Schema for `widget.group_definition`

Required:

- **layout_type** (String, Required) The layout type of the group, only 'ordered' for now.
- **widget** (Block List, Min: 1) The list of widgets in this group. (see [below for nested schema](#nestedblock--widget--group_definition--widget))

Optional:

- **title** (String, Optional) The title of the group.

<a id="nestedblock--widget--group_definition--widget"></a>
### Nested Schema for `widget.group_definition.widget`

Optional:

- **alert_graph_definition** (Block List, Max: 1) The definition for a Alert Graph widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--alert_graph_definition))
- **alert_value_definition** (Block List, Max: 1) The definition for a Alert Value widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--alert_value_definition))
- **change_definition** (Block List, Max: 1) The definition for a Change  widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--change_definition))
- **check_status_definition** (Block List, Max: 1) The definition for a Check Status widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--check_status_definition))
- **distribution_definition** (Block List, Max: 1) The definition for a Distribution widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--distribution_definition))
- **event_stream_definition** (Block List, Max: 1) The definition for a Event Stream widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--event_stream_definition))
- **event_timeline_definition** (Block List, Max: 1) The definition for a Event Timeline widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--event_timeline_definition))
- **free_text_definition** (Block List, Max: 1) The definition for a Free Text widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--free_text_definition))
- **heatmap_definition** (Block List, Max: 1) The definition for a Heatmap widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--heatmap_definition))
- **hostmap_definition** (Block List, Max: 1) The definition for a Hostmap widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--hostmap_definition))
- **iframe_definition** (Block List, Max: 1) The definition for an Iframe widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--iframe_definition))
- **image_definition** (Block List, Max: 1) The definition for an Image widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--image_definition))
- **layout** (Map of String, Optional, Deprecated) The layout of the widget on a 'free' dashboard. **Deprecated.** Define `widget_layout` list with one element instead.
- **log_stream_definition** (Block List, Max: 1) The definition for an Log Stream widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--log_stream_definition))
- **manage_status_definition** (Block List, Max: 1) The definition for an Manage Status widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--manage_status_definition))
- **note_definition** (Block List, Max: 1) The definition for a Note widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--note_definition))
- **query_table_definition** (Block List, Max: 1) The definition for a Query Table widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--query_table_definition))
- **query_value_definition** (Block List, Max: 1) The definition for a Query Value widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--query_value_definition))
- **scatterplot_definition** (Block List, Max: 1) The definition for a Scatterplot widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--scatterplot_definition))
- **service_level_objective_definition** (Block List, Max: 1) The definition for a Service Level Objective widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--service_level_objective_definition))
- **servicemap_definition** (Block List, Max: 1) The definition for a Service Map widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--servicemap_definition))
- **timeseries_definition** (Block List, Max: 1) The definition for a Timeseries widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--timeseries_definition))
- **toplist_definition** (Block List, Max: 1) The definition for a Toplist widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--toplist_definition))
- **trace_service_definition** (Block List, Max: 1) The definition for a Trace Service widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition))
- **widget_layout** (Block List, Max: 1) The layout of the widget on a 'free' dashboard. (see [below for nested schema](#nestedblock--widget--group_definition--widget--widget_layout))

Read-only:

- **id** (Number, Read-only) The ID of the widget.

<a id="nestedblock--widget--group_definition--widget--alert_graph_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **alert_id** (String, Required) The ID of the monitor used by the widget.
- **viz_type** (String, Required) Type of visualization to use when displaying the widget. Either `timeseries` or `toplist`.

Optional:

- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--group_definition--widget--alert_value_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **alert_id** (String, Required) The ID of the monitor used by the widget.

Optional:

- **precision** (Number, Optional) The precision to use when displaying the value. Use `*` for maximum precision.
- **text_align** (String, Optional) The alignment of the text in the widget.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.
- **unit** (String, Optional) The unit for the value displayed in the widget.


<a id="nestedblock--widget--group_definition--widget--change_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--custom_link))
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple request blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--group_definition--widget--id--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.id.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--group_definition--widget--id--request"></a>
### Nested Schema for `widget.group_definition.widget.id.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--apm_query))
- **change_type** (String, Optional) Whether to show absolute or relative change. One of `absolute`, `relative`.
- **compare_to** (String, Optional) Choose from when to compare current data to. One of `hour_before`, `day_before`, `week_before` or `month_before`.
- **increase_good** (Boolean, Optional) Boolean indicating whether an increase in the value is good (thus displayed in green) or not (thus displayed in red).
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--log_query))
- **order_by** (String, Optional) One of `change`, `name`, `present` (present value) or `past` (past value).
- **order_dir** (String, Optional) Either `asc` (ascending) or `desc` (descending).
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query))
- **show_present** (Boolean, Optional) If set to `true`, displays current value.

<a id="nestedblock--widget--group_definition--widget--id--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--show_present--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--show_present--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--show_present--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--show_present--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--show_present--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--show_present--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--show_present--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--show_present--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--show_present--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--show_present--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--show_present--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--show_present--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--show_present--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--show_present--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--show_present--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--show_present--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--show_present--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.show_present.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--group_definition--widget--check_status_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **check** (String, Required) The check to use in the widget.
- **grouping** (String, Required) Either `check` or `cluster`, depending on whether the widget should use a single check or a cluster of checks.

Optional:

- **group** (String, Optional) The check group to use in the widget.
- **group_by** (List of String, Optional) When `grouping = "cluster"`, indicates a list of tags to use for grouping.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **tags** (List of String, Optional) List of tags to use in the widget.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--group_definition--widget--distribution_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **legend_size** (String, Optional) The size of the legend displayed in the widget.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple request blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request))
- **show_legend** (Boolean, Optional) Whether or not to show the legend on this widget.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--group_definition--widget--id--request"></a>
### Nested Schema for `widget.group_definition.widget.id.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query))
- **style** (Block List, Max: 1) Style of the widget graph. One nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style))

<a id="nestedblock--widget--group_definition--widget--id--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--style"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Optional:

- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.




<a id="nestedblock--widget--group_definition--widget--event_stream_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **query** (String, Required) The query to use in the widget.

Optional:

- **event_size** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **tags_execution** (String, Optional) The execution method for multi-value filters. Can be either `and` or `or`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--group_definition--widget--event_timeline_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **query** (String, Required) The query to use in the widget.

Optional:

- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **tags_execution** (String, Optional) The execution method for multi-value filters. Can be either `and` or `or`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--group_definition--widget--free_text_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **text** (String, Required) The text to display in the widget.

Optional:

- **color** (String, Optional) The color of the text in the widget.
- **font_size** (String, Optional) The size of the text in the widget.
- **text_align** (String, Optional) The alignment of the text in the widget.


<a id="nestedblock--widget--group_definition--widget--heatmap_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--custom_link))
- **event** (Block List) The definition of the event to overlay on the graph. Multiple `event` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--event))
- **legend_size** (String, Optional) The size of the legend displayed in the widget.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request))
- **show_legend** (Boolean, Optional) Whether or not to show the legend on this widget.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.
- **yaxis** (Block List, Max: 1) Nested block describing the Y-Axis Controls. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--yaxis))

<a id="nestedblock--widget--group_definition--widget--id--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.id.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--group_definition--widget--id--event"></a>
### Nested Schema for `widget.group_definition.widget.id.event`

Required:

- **q** (String, Required) The event query to use in the widget.

Optional:

- **tags_execution** (String, Optional) The execution method for multi-value filters.


<a id="nestedblock--widget--group_definition--widget--id--request"></a>
### Nested Schema for `widget.group_definition.widget.id.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query))
- **style** (Block List, Max: 1) Style of the widget graph. One nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style))

<a id="nestedblock--widget--group_definition--widget--id--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--style"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Optional:

- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.



<a id="nestedblock--widget--group_definition--widget--id--yaxis"></a>
### Nested Schema for `widget.group_definition.widget.id.yaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.



<a id="nestedblock--widget--group_definition--widget--hostmap_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--custom_link))
- **group** (List of String, Optional) The list of tags to group nodes by.
- **no_group_hosts** (Boolean, Optional) Boolean indicating whether to show ungrouped nodes.
- **no_metric_hosts** (Boolean, Optional) Boolean indicating whether to show nodes with no metrics.
- **node_type** (String, Optional) The type of node used. Either `host` or `container`.
- **request** (Block List, Max: 1) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request))
- **scope** (List of String, Optional) The list of tags to filter nodes by.
- **style** (Block List, Max: 1) Style of the widget graph. One nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--style))
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--group_definition--widget--id--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.id.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--group_definition--widget--id--request"></a>
### Nested Schema for `widget.group_definition.widget.id.request`

Optional:

- **fill** (Block List) The query used to fill the map. Exactly one nested block is allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--fill))
- **size** (Block List) The query used to size the map. Exactly one nested block is allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size))

<a id="nestedblock--widget--group_definition--widget--id--request--fill"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--size--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--size--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--size--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.




<a id="nestedblock--widget--group_definition--widget--id--request--size"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--size--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--size--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--size--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--group_definition--widget--id--style"></a>
### Nested Schema for `widget.group_definition.widget.id.style`

Optional:

- **fill_max** (String, Optional) Max value to use to color the map.
- **fill_min** (String, Optional) Min value to use to color the map.
- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.
- **palette_flip** (Boolean, Optional) Boolean indicating whether to flip the palette tones.



<a id="nestedblock--widget--group_definition--widget--iframe_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **url** (String, Required) The URL to use as a data source for the widget.


<a id="nestedblock--widget--group_definition--widget--image_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **url** (String, Required) The URL to use as a data source for the widget.

Optional:

- **margin** (String, Optional) The margins to use around the image. Either `small` or `large`.
- **sizing** (String, Optional) The preferred method to adapt the dimensions of the image to those of the widget. One of `center` (center the image in the tile), `zoom` (zoom the image to cover the whole tile) or `fit` (fit the image dimensions to those of the tile).


<a id="nestedblock--widget--group_definition--widget--log_stream_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **columns** (List of String, Optional) Stringified list of columns to use. Example: `["column1","column2","column3"]`.
- **indexes** (List of String, Optional) An array of index names to query in the stream.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **logset** (String, Optional, Deprecated) ID of the logset to use. Deprecated Use `indexes` instead. **Deprecated.** This parameter has been deprecated. Use `indexes` instead.
- **message_display** (String, Optional) One of: ['inline', 'expanded-md', 'expanded-lg']
- **query** (String, Optional) The query to use in the widget.
- **show_date_column** (Boolean, Optional) If the date column should be displayed.
- **show_message_column** (Boolean, Optional) If the message column should be displayed.
- **sort** (Block List, Max: 1) The facet and order to sort the data based upon. Example: `{"column": "time", "order": "desc"}`. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--sort))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--group_definition--widget--id--sort"></a>
### Nested Schema for `widget.group_definition.widget.id.sort`

Required:

- **column** (String, Required) Facet path for the column
- **order** (String, Required) Widget sorting methods.



<a id="nestedblock--widget--group_definition--widget--manage_status_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **query** (String, Required) The query to use in the widget.

Optional:

- **color_preference** (String, Optional) Whether to colorize text or background. One of `text`, `background`.
- **count** (Number, Optional, Deprecated) The number of monitors to display. **Deprecated.** This parameter has been deprecated.
- **display_format** (String, Optional) The display setting to use. One of `counts`, `list`, or `countsAndList`.
- **hide_zero_counts** (Boolean, Optional) Boolean indicating whether to hide empty categories.
- **show_last_triggered** (Boolean, Optional) Boolean indicating whether to show when monitors/groups last triggered.
- **sort** (String, Optional) The method to use to sort monitors. Example: `status,asc`.
- **start** (Number, Optional, Deprecated) The start of the list. Typically 0. **Deprecated.** This parameter has been deprecated.
- **summary_type** (String, Optional) One of: ['monitors', 'groups', 'combined']
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--group_definition--widget--note_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **content** (String, Required) Content of the note.

Optional:

- **background_color** (String, Optional) Background color of the note.
- **font_size** (String, Optional) Size of the text.
- **show_tick** (Boolean, Optional) Whether to show a tick or not.
- **text_align** (String, Optional) The alignment of the widget's text. One of `left`, `center`, or `right`.
- **tick_edge** (String, Optional) When `tick = true`, string indicating on which side of the widget the tick should be displayed. One of `bottom`, `top`, `left`, `right`.
- **tick_pos** (String, Optional) When `tick = true`, string with a percent sign indicating the position of the tick. Example: use `tick_pos = "50%"` for centered alignment.


<a id="nestedblock--widget--group_definition--widget--query_table_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--custom_link))
- **has_search_bar** (String, Optional) Controls the display of the search bar. One of `auto`, `always`, `never`.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the `request` block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--group_definition--widget--id--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.id.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--group_definition--widget--id--request"></a>
### Nested Schema for `widget.group_definition.widget.id.request`

Optional:

- **aggregator** (String, Optional) The aggregator to use for time aggregation. One of `avg`, `min`, `max`, `sum`, `last`.
- **alias** (String, Optional) The alias for the column name. Default is the metric name.
- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--apm_query))
- **apm_stats_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--apm_stats_query))
- **cell_display_mode** (List of String, Optional) A list of display modes for each table cell. List items one of `number`, `bar`.
- **conditional_formats** (Block List) Conditional formats allow you to set the color of your widget content or background, depending on a rule applied to your data. Multiple `conditional_formats` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--conditional_formats))
- **limit** (Number, Optional) The number of lines to show in the table.
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--log_query))
- **order** (String, Optional) The sort order for the rows. One of `desc` or `asc`.
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query))

<a id="nestedblock--widget--group_definition--widget--id--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--apm_stats_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **env** (String, Required) Environment name.
- **name** (String, Required) Operation name associated with service.
- **primary_tag** (String, Required) The organization's host group name and value.
- **row_type** (String, Required) The level of detail for the request.
- **service** (String, Required) Service name.

Optional:

- **columns** (Block List) Column properties used by the front end for display. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--columns))
- **resource** (String, Optional) Resource name.

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--columns"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.resource`

Required:

- **name** (String, Required) Column name.

Optional:

- **alias** (String, Optional) A user-assigned alias for the column.
- **cell_display_mode** (String, Optional) A list of display modes for each table cell.
- **order** (String, Optional) Widget sorting methods.



<a id="nestedblock--widget--group_definition--widget--id--request--conditional_formats"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **comparator** (String, Required) Comparator to use. One of `>`, `>=`, `<`, or `<=`.
- **palette** (String, Required) Color palette to apply. One of `blue`, `custom_bg`, `custom_image`, `custom_text`, `gray_on_white`, `grey`, `green`, `orange`, `red`, `red_on_white`, `white_on_gray`, `white_on_green`, `green_on_white`, `white_on_red`, `white_on_yellow`, `yellow_on_white`, `black_on_light_yellow`, `black_on_light_green` or `black_on_light_red`.
- **value** (Number, Required) Value for the comparator.

Optional:

- **custom_bg_color** (String, Optional) Color palette to apply to the background, same values available as palette.
- **custom_fg_color** (String, Optional) Color palette to apply to the foreground, same values available as palette.
- **hide_value** (Boolean, Optional) Setting this to True hides values.
- **image_url** (String, Optional) Displays an image as the background.
- **metric** (String, Optional) Metric from the request to correlate this conditional format with.
- **timeframe** (String, Optional) Defines the displayed timeframe.


<a id="nestedblock--widget--group_definition--widget--id--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--group_definition--widget--query_value_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **autoscale** (Boolean, Optional) Boolean indicating whether to automatically scale the tile.
- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--custom_link))
- **custom_unit** (String, Optional) The unit for the value displayed in the widget.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **precision** (Number, Optional) The precision to use when displaying the tile.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the `request` block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request))
- **text_align** (String, Optional) The alignment of the widget's text. One of `left`, `center`, or `right`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--group_definition--widget--id--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.id.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--group_definition--widget--id--request"></a>
### Nested Schema for `widget.group_definition.widget.id.request`

Optional:

- **aggregator** (String, Optional) The aggregator to use for time aggregation. One of `avg`, `min`, `max`, `sum`, `last`.
- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--apm_query))
- **conditional_formats** (Block List) Conditional formats allow you to set the color of your widget content or background, depending on a rule applied to your data. Multiple `conditional_formats` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--conditional_formats))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query))

<a id="nestedblock--widget--group_definition--widget--id--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--conditional_formats"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **comparator** (String, Required) Comparator to use. One of `>`, `>=`, `<`, or `<=`.
- **palette** (String, Required) Color palette to apply. One of `blue`, `custom_bg`, `custom_image`, `custom_text`, `gray_on_white`, `grey`, `green`, `orange`, `red`, `red_on_white`, `white_on_gray`, `white_on_green`, `green_on_white`, `white_on_red`, `white_on_yellow`, `yellow_on_white`, `black_on_light_yellow`, `black_on_light_green` or `black_on_light_red`.
- **value** (Number, Required) Value for the comparator.

Optional:

- **custom_bg_color** (String, Optional) Color palette to apply to the background, same values available as palette.
- **custom_fg_color** (String, Optional) Color palette to apply to the foreground, same values available as palette.
- **hide_value** (Boolean, Optional) Setting this to True hides values.
- **image_url** (String, Optional) Displays an image as the background.
- **metric** (String, Optional) Metric from the request to correlate this conditional format with.
- **timeframe** (String, Optional) Defines the displayed timeframe.


<a id="nestedblock--widget--group_definition--widget--id--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--group_definition--widget--scatterplot_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **color_by_groups** (List of String, Optional) List of groups used for colors.
- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--custom_link))
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List, Max: 1) Nested block describing the request to use when displaying the widget. Exactly one `request` block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.
- **xaxis** (Block List, Max: 1) Nested block describing the X-Axis Controls. Exactly one nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--xaxis))
- **yaxis** (Block List, Max: 1) Nested block describing the Y-Axis Controls. Exactly one nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--yaxis))

<a id="nestedblock--widget--group_definition--widget--id--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.id.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--group_definition--widget--id--request"></a>
### Nested Schema for `widget.group_definition.widget.id.request`

Optional:

- **x** (Block List) The query used for the X-Axis. Exactly one nested block is allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--x))
- **y** (Block List) The query used for the Y-Axis. Exactly one nested block is allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y))

<a id="nestedblock--widget--group_definition--widget--id--request--x"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y`

Optional:

- **aggregator** (String, Optional) Aggregator used for the request.
- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--y--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--y--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--y--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.




<a id="nestedblock--widget--group_definition--widget--id--request--y"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y`

Optional:

- **aggregator** (String, Optional) Aggregator used for the request.
- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--y--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--y--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--y--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--group_definition--widget--id--xaxis"></a>
### Nested Schema for `widget.group_definition.widget.id.xaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.


<a id="nestedblock--widget--group_definition--widget--id--yaxis"></a>
### Nested Schema for `widget.group_definition.widget.id.yaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.



<a id="nestedblock--widget--group_definition--widget--service_level_objective_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **slo_id** (String, Required) The ID of the service level objective used by the widget.
- **time_windows** (List of String, Required) List of time windows to display in the widget. Each value in the list must be one of `7d`, `30d`, `90d`, `week_to_date`, `previous_week`, `month_to_date`, or `previous_month`.
- **view_mode** (String, Required) View mode for the widget. One of `overall`, `component`, or `both`.
- **view_type** (String, Required) Type of view to use when displaying the widget. Only `detail` is currently supported.

Optional:

- **show_error_budget** (Boolean, Optional) Whether to show the error budget or not.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--group_definition--widget--servicemap_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **filters** (List of String, Required) Your environment and primary tag (or `*` if enabled for your account).
- **service** (String, Required) The ID of the service you want to map.

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--custom_link))
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--group_definition--widget--id--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.id.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.



<a id="nestedblock--widget--group_definition--widget--timeseries_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--custom_link))
- **event** (Block List) The definition of the event to overlay on the graph. Multiple `event` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--event))
- **legend_size** (String, Optional) The size of the legend displayed in the widget.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **marker** (Block List) Nested block describing the marker to use when displaying the widget. The structure of this block is described below. Multiple `marker` blocks are allowed within a given `tile_def` block. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--marker))
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `network_query`, `security_query` or `process_query` is required within the `request` block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request))
- **right_yaxis** (Block List, Max: 1) Nested block describing the right Y-Axis Controls. See the `on_right_yaxis` property for which request will use this axis. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--right_yaxis))
- **show_legend** (Boolean, Optional) Whether or not to show the legend on this widget.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.
- **yaxis** (Block List, Max: 1) Nested block describing the Y-Axis Controls. The structure of this block is described below (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--yaxis))

<a id="nestedblock--widget--group_definition--widget--id--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.id.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--group_definition--widget--id--event"></a>
### Nested Schema for `widget.group_definition.widget.id.event`

Required:

- **q** (String, Required) The event query to use in the widget.

Optional:

- **tags_execution** (String, Optional) The execution method for multi-value filters.


<a id="nestedblock--widget--group_definition--widget--id--marker"></a>
### Nested Schema for `widget.group_definition.widget.id.marker`

Required:

- **value** (String, Required) Mathematical expression describing the marker. Examples: `y > 1`, `-5 < y < 0`, `y = 19`.

Optional:

- **display_type** (String, Optional) How the marker lines will look. Possible values are one of {`error`, `warning`, `info`, `ok`} combined with one of {`dashed`, `solid`, `bold`}. Example: `error dashed`.
- **label** (String, Optional) A label for the line or range.


<a id="nestedblock--widget--group_definition--widget--id--request"></a>
### Nested Schema for `widget.group_definition.widget.id.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--apm_query))
- **display_type** (String, Optional) How the marker lines will look. Possible values are one of {`error`, `warning`, `info`, `ok`} combined with one of {`dashed`, `solid`, `bold`}. Example: `error dashed`.
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--log_query))
- **metadata** (Block List) Used to define expression aliases. Multiple `metadata` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--metadata))
- **network_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--network_query))
- **on_right_yaxis** (Boolean, Optional) Boolean indicating whether the request will use the right or left Y-Axis.
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query))
- **style** (Block List, Max: 1) Style of the widget graph. Exactly one `style` block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style))

<a id="nestedblock--widget--group_definition--widget--id--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--metadata"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **expression** (String, Required) Expression name.

Optional:

- **alias_name** (String, Optional) Expression alias.


<a id="nestedblock--widget--group_definition--widget--id--request--network_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--style"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Optional:

- **line_type** (String, Optional) Type of lines displayed. Available values are: `dashed`, `dotted`, or `solid`.
- **line_width** (String, Optional) Width of line displayed. Available values are: `normal`, `thick`, or `thin`.
- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.



<a id="nestedblock--widget--group_definition--widget--id--right_yaxis"></a>
### Nested Schema for `widget.group_definition.widget.id.right_yaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.


<a id="nestedblock--widget--group_definition--widget--id--yaxis"></a>
### Nested Schema for `widget.group_definition.widget.id.yaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.



<a id="nestedblock--widget--group_definition--widget--toplist_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--custom_link))
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the `request` block). (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--group_definition--widget--id--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.id.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--group_definition--widget--id--request"></a>
### Nested Schema for `widget.group_definition.widget.id.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--apm_query))
- **conditional_formats** (Block List) Conditional formats allow you to set the color of your widget content or background, depending on a rule applied to your data. Multiple `conditional_formats` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--conditional_formats))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--security_query))
- **style** (Block List, Max: 1) Define request widget style. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style))

<a id="nestedblock--widget--group_definition--widget--id--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--conditional_formats"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **comparator** (String, Required) Comparator to use. One of `>`, `>=`, `<`, or `<=`.
- **palette** (String, Required) Color palette to apply. One of `blue`, `custom_bg`, `custom_image`, `custom_text`, `gray_on_white`, `grey`, `green`, `orange`, `red`, `red_on_white`, `white_on_gray`, `white_on_green`, `green_on_white`, `white_on_red`, `white_on_yellow`, `yellow_on_white`, `black_on_light_yellow`, `black_on_light_green` or `black_on_light_red`.
- **value** (Number, Required) Value for the comparator.

Optional:

- **custom_bg_color** (String, Optional) Color palette to apply to the background, same values available as palette.
- **custom_fg_color** (String, Optional) Color palette to apply to the foreground, same values available as palette.
- **hide_value** (Boolean, Optional) Setting this to True hides values.
- **image_url** (String, Optional) Displays an image as the background.
- **metric** (String, Optional) Metric from the request to correlate this conditional format with.
- **timeframe** (String, Optional) Defines the displayed timeframe.


<a id="nestedblock--widget--group_definition--widget--id--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--group_definition--widget--id--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--group_definition--widget--id--request--style--compute_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--group_definition--widget--id--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query))

<a id="nestedblock--widget--group_definition--widget--id--request--style--search_query--sort_query"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--group_definition--widget--id--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--group_definition--widget--id--request--style"></a>
### Nested Schema for `widget.group_definition.widget.id.request.style`

Optional:

- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.




<a id="nestedblock--widget--group_definition--widget--trace_service_definition"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **env** (String, Required) APM environment.
- **service** (String, Required) APM service.
- **span_name** (String, Required) APM span name

Optional:

- **display_format** (String, Optional) Number of columns to display. Available values are: `one_column`, `two_column`, or `three_column`.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **show_breakdown** (Boolean, Optional) Whether to show the latency breakdown or not.
- **show_distribution** (Boolean, Optional) Whether to show the latency distribution or not.
- **show_errors** (Boolean, Optional) Whether to show the error metrics or not.
- **show_hits** (Boolean, Optional) Whether to show the hits metrics or not
- **show_latency** (Boolean, Optional) Whether to show the latency metrics or not.
- **show_resource_list** (Boolean, Optional) Whether to show the resource list or not.
- **size_format** (String, Optional) Size of the widget. Available values are: `small`, `medium`, or `large`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--group_definition--widget--widget_layout"></a>
### Nested Schema for `widget.group_definition.widget.id`

Required:

- **height** (Number, Required) The height of the widget.
- **width** (Number, Required) The width of the widget.
- **x** (Number, Required) The position of the widget on the x (horizontal) axis. Should be greater or equal to 0.
- **y** (Number, Required) The position of the widget on the y (vertical) axis. Should be greater or equal to 0.




<a id="nestedblock--widget--heatmap_definition"></a>
### Nested Schema for `widget.heatmap_definition`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--custom_link))
- **event** (Block List) The definition of the event to overlay on the graph. Multiple `event` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--event))
- **legend_size** (String, Optional) The size of the legend displayed in the widget.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--heatmap_definition--request))
- **show_legend** (Boolean, Optional) Whether or not to show the legend on this widget.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.
- **yaxis** (Block List, Max: 1) Nested block describing the Y-Axis Controls. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--yaxis))

<a id="nestedblock--widget--heatmap_definition--custom_link"></a>
### Nested Schema for `widget.heatmap_definition.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--heatmap_definition--event"></a>
### Nested Schema for `widget.heatmap_definition.event`

Required:

- **q** (String, Required) The event query to use in the widget.

Optional:

- **tags_execution** (String, Optional) The execution method for multi-value filters.


<a id="nestedblock--widget--heatmap_definition--request"></a>
### Nested Schema for `widget.heatmap_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--security_query))
- **style** (Block List, Max: 1) Style of the widget graph. One nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style))

<a id="nestedblock--widget--heatmap_definition--request--apm_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--heatmap_definition--request--style--compute_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--heatmap_definition--request--style--group_by"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--heatmap_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--heatmap_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.heatmap_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--heatmap_definition--request--log_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--heatmap_definition--request--style--compute_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--heatmap_definition--request--style--group_by"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--heatmap_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--heatmap_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.heatmap_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--heatmap_definition--request--process_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--heatmap_definition--request--rum_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--heatmap_definition--request--style--compute_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--heatmap_definition--request--style--group_by"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--heatmap_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--heatmap_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.heatmap_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--heatmap_definition--request--security_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--heatmap_definition--request--style--compute_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--heatmap_definition--request--style--group_by"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--heatmap_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--heatmap_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.heatmap_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--heatmap_definition--request--style"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Optional:

- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.



<a id="nestedblock--widget--heatmap_definition--yaxis"></a>
### Nested Schema for `widget.heatmap_definition.yaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.



<a id="nestedblock--widget--hostmap_definition"></a>
### Nested Schema for `widget.hostmap_definition`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--custom_link))
- **group** (List of String, Optional) The list of tags to group nodes by.
- **no_group_hosts** (Boolean, Optional) Boolean indicating whether to show ungrouped nodes.
- **no_metric_hosts** (Boolean, Optional) Boolean indicating whether to show nodes with no metrics.
- **node_type** (String, Optional) The type of node used. Either `host` or `container`.
- **request** (Block List, Max: 1) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request))
- **scope** (List of String, Optional) The list of tags to filter nodes by.
- **style** (Block List, Max: 1) Style of the widget graph. One nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--style))
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--hostmap_definition--custom_link"></a>
### Nested Schema for `widget.hostmap_definition.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--hostmap_definition--request"></a>
### Nested Schema for `widget.hostmap_definition.request`

Optional:

- **fill** (Block List) The query used to fill the map. Exactly one nested block is allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--fill))
- **size** (Block List) The query used to size the map. Exactly one nested block is allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block). (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size))

<a id="nestedblock--widget--hostmap_definition--request--fill"></a>
### Nested Schema for `widget.hostmap_definition.request.size`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query))

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--compute_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--search_query--sort_query))

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--search_query--sort_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--hostmap_definition--request--size--log_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--hostmap_definition--request--size--log_query--compute_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--hostmap_definition--request--size--log_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--search_query--sort_query))

<a id="nestedblock--widget--hostmap_definition--request--size--log_query--search_query--sort_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--hostmap_definition--request--size--log_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--hostmap_definition--request--size--process_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.process_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--hostmap_definition--request--size--rum_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--compute_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--search_query--sort_query))

<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--search_query--sort_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--hostmap_definition--request--size--security_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--hostmap_definition--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--hostmap_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--hostmap_definition--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--hostmap_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.




<a id="nestedblock--widget--hostmap_definition--request--size"></a>
### Nested Schema for `widget.hostmap_definition.request.size`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query))

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--compute_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--search_query--sort_query))

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--search_query--sort_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--hostmap_definition--request--size--log_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--hostmap_definition--request--size--log_query--compute_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--hostmap_definition--request--size--log_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--search_query--sort_query))

<a id="nestedblock--widget--hostmap_definition--request--size--log_query--search_query--sort_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--hostmap_definition--request--size--log_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--hostmap_definition--request--size--process_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.process_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--hostmap_definition--request--size--rum_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--compute_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--search_query--sort_query))

<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--search_query--sort_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--hostmap_definition--request--size--security_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--hostmap_definition--request--size--security_query--compute_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--hostmap_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--search_query--sort_query))

<a id="nestedblock--widget--hostmap_definition--request--size--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--hostmap_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--hostmap_definition--style"></a>
### Nested Schema for `widget.hostmap_definition.style`

Optional:

- **fill_max** (String, Optional) Max value to use to color the map.
- **fill_min** (String, Optional) Min value to use to color the map.
- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.
- **palette_flip** (Boolean, Optional) Boolean indicating whether to flip the palette tones.



<a id="nestedblock--widget--iframe_definition"></a>
### Nested Schema for `widget.iframe_definition`

Required:

- **url** (String, Required) The URL to use as a data source for the widget.


<a id="nestedblock--widget--image_definition"></a>
### Nested Schema for `widget.image_definition`

Required:

- **url** (String, Required) The URL to use as a data source for the widget.

Optional:

- **margin** (String, Optional) The margins to use around the image. Either `small` or `large`.
- **sizing** (String, Optional) The preferred method to adapt the dimensions of the image to those of the widget. One of `center` (center the image in the tile), `zoom` (zoom the image to cover the whole tile) or `fit` (fit the image dimensions to those of the tile).


<a id="nestedblock--widget--log_stream_definition"></a>
### Nested Schema for `widget.log_stream_definition`

Optional:

- **columns** (List of String, Optional) Stringified list of columns to use. Example: `["column1","column2","column3"]`.
- **indexes** (List of String, Optional) An array of index names to query in the stream.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **logset** (String, Optional, Deprecated) ID of the logset to use. Deprecated Use `indexes` instead. **Deprecated.** This parameter has been deprecated. Use `indexes` instead.
- **message_display** (String, Optional) One of: ['inline', 'expanded-md', 'expanded-lg']
- **query** (String, Optional) The query to use in the widget.
- **show_date_column** (Boolean, Optional) If the date column should be displayed.
- **show_message_column** (Boolean, Optional) If the message column should be displayed.
- **sort** (Block List, Max: 1) The facet and order to sort the data based upon. Example: `{"column": "time", "order": "desc"}`. (see [below for nested schema](#nestedblock--widget--log_stream_definition--sort))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--log_stream_definition--sort"></a>
### Nested Schema for `widget.log_stream_definition.sort`

Required:

- **column** (String, Required) Facet path for the column
- **order** (String, Required) Widget sorting methods.



<a id="nestedblock--widget--manage_status_definition"></a>
### Nested Schema for `widget.manage_status_definition`

Required:

- **query** (String, Required) The query to use in the widget.

Optional:

- **color_preference** (String, Optional) Whether to colorize text or background. One of `text`, `background`.
- **count** (Number, Optional, Deprecated) The number of monitors to display. **Deprecated.** This parameter has been deprecated.
- **display_format** (String, Optional) The display setting to use. One of `counts`, `list`, or `countsAndList`.
- **hide_zero_counts** (Boolean, Optional) Boolean indicating whether to hide empty categories.
- **show_last_triggered** (Boolean, Optional) Boolean indicating whether to show when monitors/groups last triggered.
- **sort** (String, Optional) The method to use to sort monitors. Example: `status,asc`.
- **start** (Number, Optional, Deprecated) The start of the list. Typically 0. **Deprecated.** This parameter has been deprecated.
- **summary_type** (String, Optional) One of: ['monitors', 'groups', 'combined']
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--note_definition"></a>
### Nested Schema for `widget.note_definition`

Required:

- **content** (String, Required) Content of the note.

Optional:

- **background_color** (String, Optional) Background color of the note.
- **font_size** (String, Optional) Size of the text.
- **show_tick** (Boolean, Optional) Whether to show a tick or not.
- **text_align** (String, Optional) The alignment of the widget's text. One of `left`, `center`, or `right`.
- **tick_edge** (String, Optional) When `tick = true`, string indicating on which side of the widget the tick should be displayed. One of `bottom`, `top`, `left`, `right`.
- **tick_pos** (String, Optional) When `tick = true`, string with a percent sign indicating the position of the tick. Example: use `tick_pos = "50%"` for centered alignment.


<a id="nestedblock--widget--query_table_definition"></a>
### Nested Schema for `widget.query_table_definition`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--custom_link))
- **has_search_bar** (String, Optional) Controls the display of the search bar. One of `auto`, `always`, `never`.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the `request` block). (see [below for nested schema](#nestedblock--widget--query_table_definition--request))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--query_table_definition--custom_link"></a>
### Nested Schema for `widget.query_table_definition.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--query_table_definition--request"></a>
### Nested Schema for `widget.query_table_definition.request`

Optional:

- **aggregator** (String, Optional) The aggregator to use for time aggregation. One of `avg`, `min`, `max`, `sum`, `last`.
- **alias** (String, Optional) The alias for the column name. Default is the metric name.
- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--apm_query))
- **apm_stats_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--apm_stats_query))
- **cell_display_mode** (List of String, Optional) A list of display modes for each table cell. List items one of `number`, `bar`.
- **conditional_formats** (Block List) Conditional formats allow you to set the color of your widget content or background, depending on a rule applied to your data. Multiple `conditional_formats` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--conditional_formats))
- **limit** (Number, Optional) The number of lines to show in the table.
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--log_query))
- **order** (String, Optional) The sort order for the rows. One of `desc` or `asc`.
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query))

<a id="nestedblock--widget--query_table_definition--request--apm_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--query_table_definition--request--security_query--compute_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--query_table_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by--sort_query))

<a id="nestedblock--widget--query_table_definition--request--security_query--group_by--sort_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--query_table_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--query_table_definition--request--apm_stats_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **env** (String, Required) Environment name.
- **name** (String, Required) Operation name associated with service.
- **primary_tag** (String, Required) The organization's host group name and value.
- **row_type** (String, Required) The level of detail for the request.
- **service** (String, Required) Service name.

Optional:

- **columns** (Block List) Column properties used by the front end for display. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--columns))
- **resource** (String, Optional) Resource name.

<a id="nestedblock--widget--query_table_definition--request--security_query--columns"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.columns`

Required:

- **name** (String, Required) Column name.

Optional:

- **alias** (String, Optional) A user-assigned alias for the column.
- **cell_display_mode** (String, Optional) A list of display modes for each table cell.
- **order** (String, Optional) Widget sorting methods.



<a id="nestedblock--widget--query_table_definition--request--conditional_formats"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **comparator** (String, Required) Comparator to use. One of `>`, `>=`, `<`, or `<=`.
- **palette** (String, Required) Color palette to apply. One of `blue`, `custom_bg`, `custom_image`, `custom_text`, `gray_on_white`, `grey`, `green`, `orange`, `red`, `red_on_white`, `white_on_gray`, `white_on_green`, `green_on_white`, `white_on_red`, `white_on_yellow`, `yellow_on_white`, `black_on_light_yellow`, `black_on_light_green` or `black_on_light_red`.
- **value** (Number, Required) Value for the comparator.

Optional:

- **custom_bg_color** (String, Optional) Color palette to apply to the background, same values available as palette.
- **custom_fg_color** (String, Optional) Color palette to apply to the foreground, same values available as palette.
- **hide_value** (Boolean, Optional) Setting this to True hides values.
- **image_url** (String, Optional) Displays an image as the background.
- **metric** (String, Optional) Metric from the request to correlate this conditional format with.
- **timeframe** (String, Optional) Defines the displayed timeframe.


<a id="nestedblock--widget--query_table_definition--request--log_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--query_table_definition--request--security_query--compute_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--query_table_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by--sort_query))

<a id="nestedblock--widget--query_table_definition--request--security_query--group_by--sort_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--query_table_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--query_table_definition--request--process_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--query_table_definition--request--rum_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--query_table_definition--request--security_query--compute_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--query_table_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by--sort_query))

<a id="nestedblock--widget--query_table_definition--request--security_query--group_by--sort_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--query_table_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--query_table_definition--request--security_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--query_table_definition--request--security_query--compute_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--query_table_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by--sort_query))

<a id="nestedblock--widget--query_table_definition--request--security_query--group_by--sort_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--query_table_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--query_value_definition"></a>
### Nested Schema for `widget.query_value_definition`

Optional:

- **autoscale** (Boolean, Optional) Boolean indicating whether to automatically scale the tile.
- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--custom_link))
- **custom_unit** (String, Optional) The unit for the value displayed in the widget.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **precision** (Number, Optional) The precision to use when displaying the tile.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the `request` block). (see [below for nested schema](#nestedblock--widget--query_value_definition--request))
- **text_align** (String, Optional) The alignment of the widget's text. One of `left`, `center`, or `right`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--query_value_definition--custom_link"></a>
### Nested Schema for `widget.query_value_definition.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--query_value_definition--request"></a>
### Nested Schema for `widget.query_value_definition.request`

Optional:

- **aggregator** (String, Optional) The aggregator to use for time aggregation. One of `avg`, `min`, `max`, `sum`, `last`.
- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--apm_query))
- **conditional_formats** (Block List) Conditional formats allow you to set the color of your widget content or background, depending on a rule applied to your data. Multiple `conditional_formats` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--conditional_formats))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query))

<a id="nestedblock--widget--query_value_definition--request--apm_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--query_value_definition--request--security_query--compute_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--query_value_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by--sort_query))

<a id="nestedblock--widget--query_value_definition--request--security_query--group_by--sort_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--query_value_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--query_value_definition--request--conditional_formats"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **comparator** (String, Required) Comparator to use. One of `>`, `>=`, `<`, or `<=`.
- **palette** (String, Required) Color palette to apply. One of `blue`, `custom_bg`, `custom_image`, `custom_text`, `gray_on_white`, `grey`, `green`, `orange`, `red`, `red_on_white`, `white_on_gray`, `white_on_green`, `green_on_white`, `white_on_red`, `white_on_yellow`, `yellow_on_white`, `black_on_light_yellow`, `black_on_light_green` or `black_on_light_red`.
- **value** (Number, Required) Value for the comparator.

Optional:

- **custom_bg_color** (String, Optional) Color palette to apply to the background, same values available as palette.
- **custom_fg_color** (String, Optional) Color palette to apply to the foreground, same values available as palette.
- **hide_value** (Boolean, Optional) Setting this to True hides values.
- **image_url** (String, Optional) Displays an image as the background.
- **metric** (String, Optional) Metric from the request to correlate this conditional format with.
- **timeframe** (String, Optional) Defines the displayed timeframe.


<a id="nestedblock--widget--query_value_definition--request--log_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--query_value_definition--request--security_query--compute_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--query_value_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by--sort_query))

<a id="nestedblock--widget--query_value_definition--request--security_query--group_by--sort_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--query_value_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--query_value_definition--request--process_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--query_value_definition--request--rum_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--query_value_definition--request--security_query--compute_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--query_value_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by--sort_query))

<a id="nestedblock--widget--query_value_definition--request--security_query--group_by--sort_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--query_value_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--query_value_definition--request--security_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--query_value_definition--request--security_query--compute_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--query_value_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by--sort_query))

<a id="nestedblock--widget--query_value_definition--request--security_query--group_by--sort_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--query_value_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--scatterplot_definition"></a>
### Nested Schema for `widget.scatterplot_definition`

Optional:

- **color_by_groups** (List of String, Optional) List of groups used for colors.
- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--custom_link))
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List, Max: 1) Nested block describing the request to use when displaying the widget. Exactly one `request` block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.
- **xaxis** (Block List, Max: 1) Nested block describing the X-Axis Controls. Exactly one nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--xaxis))
- **yaxis** (Block List, Max: 1) Nested block describing the Y-Axis Controls. Exactly one nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--yaxis))

<a id="nestedblock--widget--scatterplot_definition--custom_link"></a>
### Nested Schema for `widget.scatterplot_definition.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--scatterplot_definition--request"></a>
### Nested Schema for `widget.scatterplot_definition.request`

Optional:

- **x** (Block List) The query used for the X-Axis. Exactly one nested block is allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the block). (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--x))
- **y** (Block List) The query used for the Y-Axis. Exactly one nested block is allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the block). (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y))

<a id="nestedblock--widget--scatterplot_definition--request--x"></a>
### Nested Schema for `widget.scatterplot_definition.request.y`

Optional:

- **aggregator** (String, Optional) Aggregator used for the request.
- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--compute_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--search_query--sort_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--search_query--sort_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--scatterplot_definition--request--y--log_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--compute_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--search_query--sort_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--search_query--sort_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--scatterplot_definition--request--y--process_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.process_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--compute_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--search_query--sort_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--search_query--sort_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--scatterplot_definition--request--y--security_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.




<a id="nestedblock--widget--scatterplot_definition--request--y"></a>
### Nested Schema for `widget.scatterplot_definition.request.y`

Optional:

- **aggregator** (String, Optional) Aggregator used for the request.
- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--compute_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--search_query--sort_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--search_query--sort_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--scatterplot_definition--request--y--log_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--compute_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--search_query--sort_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--search_query--sort_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--scatterplot_definition--request--y--process_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.process_query`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--compute_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--search_query--sort_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--search_query--sort_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--scatterplot_definition--request--y--security_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--compute_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search_query`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--search_query--sort_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--search_query--sort_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search_query.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.





<a id="nestedblock--widget--scatterplot_definition--xaxis"></a>
### Nested Schema for `widget.scatterplot_definition.xaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.


<a id="nestedblock--widget--scatterplot_definition--yaxis"></a>
### Nested Schema for `widget.scatterplot_definition.yaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.



<a id="nestedblock--widget--service_level_objective_definition"></a>
### Nested Schema for `widget.service_level_objective_definition`

Required:

- **slo_id** (String, Required) The ID of the service level objective used by the widget.
- **time_windows** (List of String, Required) List of time windows to display in the widget. Each value in the list must be one of `7d`, `30d`, `90d`, `week_to_date`, `previous_week`, `month_to_date`, or `previous_month`.
- **view_mode** (String, Required) View mode for the widget. One of `overall`, `component`, or `both`.
- **view_type** (String, Required) Type of view to use when displaying the widget. Only `detail` is currently supported.

Optional:

- **show_error_budget** (Boolean, Optional) Whether to show the error budget or not.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--servicemap_definition"></a>
### Nested Schema for `widget.servicemap_definition`

Required:

- **filters** (List of String, Required) Your environment and primary tag (or `*` if enabled for your account).
- **service** (String, Required) The ID of the service you want to map.

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--servicemap_definition--custom_link))
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--servicemap_definition--custom_link"></a>
### Nested Schema for `widget.servicemap_definition.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.



<a id="nestedblock--widget--timeseries_definition"></a>
### Nested Schema for `widget.timeseries_definition`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--custom_link))
- **event** (Block List) The definition of the event to overlay on the graph. Multiple `event` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--event))
- **legend_size** (String, Optional) The size of the legend displayed in the widget.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **marker** (Block List) Nested block describing the marker to use when displaying the widget. The structure of this block is described below. Multiple `marker` blocks are allowed within a given `tile_def` block. (see [below for nested schema](#nestedblock--widget--timeseries_definition--marker))
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `network_query`, `security_query` or `process_query` is required within the `request` block). (see [below for nested schema](#nestedblock--widget--timeseries_definition--request))
- **right_yaxis** (Block List, Max: 1) Nested block describing the right Y-Axis Controls. See the `on_right_yaxis` property for which request will use this axis. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--right_yaxis))
- **show_legend** (Boolean, Optional) Whether or not to show the legend on this widget.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.
- **yaxis** (Block List, Max: 1) Nested block describing the Y-Axis Controls. The structure of this block is described below (see [below for nested schema](#nestedblock--widget--timeseries_definition--yaxis))

<a id="nestedblock--widget--timeseries_definition--custom_link"></a>
### Nested Schema for `widget.timeseries_definition.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--timeseries_definition--event"></a>
### Nested Schema for `widget.timeseries_definition.event`

Required:

- **q** (String, Required) The event query to use in the widget.

Optional:

- **tags_execution** (String, Optional) The execution method for multi-value filters.


<a id="nestedblock--widget--timeseries_definition--marker"></a>
### Nested Schema for `widget.timeseries_definition.marker`

Required:

- **value** (String, Required) Mathematical expression describing the marker. Examples: `y > 1`, `-5 < y < 0`, `y = 19`.

Optional:

- **display_type** (String, Optional) How the marker lines will look. Possible values are one of {`error`, `warning`, `info`, `ok`} combined with one of {`dashed`, `solid`, `bold`}. Example: `error dashed`.
- **label** (String, Optional) A label for the line or range.


<a id="nestedblock--widget--timeseries_definition--request"></a>
### Nested Schema for `widget.timeseries_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--apm_query))
- **display_type** (String, Optional) How the marker lines will look. Possible values are one of {`error`, `warning`, `info`, `ok`} combined with one of {`dashed`, `solid`, `bold`}. Example: `error dashed`.
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--log_query))
- **metadata** (Block List) Used to define expression aliases. Multiple `metadata` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--metadata))
- **network_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--network_query))
- **on_right_yaxis** (Boolean, Optional) Boolean indicating whether the request will use the right or left Y-Axis.
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--security_query))
- **style** (Block List, Max: 1) Style of the widget graph. Exactly one `style` block is allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style))

<a id="nestedblock--widget--timeseries_definition--request--apm_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--timeseries_definition--request--style--compute_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--timeseries_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--timeseries_definition--request--log_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--timeseries_definition--request--style--compute_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--timeseries_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--timeseries_definition--request--metadata"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **expression** (String, Required) Expression name.

Optional:

- **alias_name** (String, Optional) Expression alias.


<a id="nestedblock--widget--timeseries_definition--request--network_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--timeseries_definition--request--style--compute_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--timeseries_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--timeseries_definition--request--process_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--timeseries_definition--request--rum_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--timeseries_definition--request--style--compute_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--timeseries_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--timeseries_definition--request--security_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--timeseries_definition--request--style--compute_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--timeseries_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--timeseries_definition--request--style"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Optional:

- **line_type** (String, Optional) Type of lines displayed. Available values are: `dashed`, `dotted`, or `solid`.
- **line_width** (String, Optional) Width of line displayed. Available values are: `normal`, `thick`, or `thin`.
- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.



<a id="nestedblock--widget--timeseries_definition--right_yaxis"></a>
### Nested Schema for `widget.timeseries_definition.right_yaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.


<a id="nestedblock--widget--timeseries_definition--yaxis"></a>
### Nested Schema for `widget.timeseries_definition.yaxis`

Optional:

- **include_zero** (Boolean, Optional) Always include zero or fit the axis to the data range.
- **label** (String, Optional) The label of the axis to display on the graph.
- **max** (String, Optional) Specify the maximum value to show on the Y-axis.
- **min** (String, Optional) Specify the minimum value to show on the Y-axis.
- **scale** (String, Optional) Specifies the scale type. One of `linear`, `log`, `pow`, `sqrt`.



<a id="nestedblock--widget--toplist_definition"></a>
### Nested Schema for `widget.toplist_definition`

Optional:

- **custom_link** (Block List) Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--custom_link))
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **request** (Block List) Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the `request` block). (see [below for nested schema](#nestedblock--widget--toplist_definition--request))
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.

<a id="nestedblock--widget--toplist_definition--custom_link"></a>
### Nested Schema for `widget.toplist_definition.custom_link`

Required:

- **label** (String, Required) The label for the custom link URL.
- **link** (String, Required) The URL of the custom link.


<a id="nestedblock--widget--toplist_definition--request"></a>
### Nested Schema for `widget.toplist_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--apm_query))
- **conditional_formats** (Block List) Conditional formats allow you to set the color of your widget content or background, depending on a rule applied to your data. Multiple `conditional_formats` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--conditional_formats))
- **log_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--log_query))
- **process_query** (Block List, Max: 1) The process query to use in the widget. The structure of this block is described below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--process_query))
- **q** (String, Optional) The metric query to use for this widget.
- **rum_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--rum_query))
- **security_query** (Block List, Max: 1) The query to use for this widget. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--security_query))
- **style** (Block List, Max: 1) Define request widget style. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style))

<a id="nestedblock--widget--toplist_definition--request--apm_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--toplist_definition--request--style--compute_query"></a>
### Nested Schema for `widget.toplist_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--toplist_definition--request--style--group_by"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--toplist_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--toplist_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.toplist_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--toplist_definition--request--conditional_formats"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **comparator** (String, Required) Comparator to use. One of `>`, `>=`, `<`, or `<=`.
- **palette** (String, Required) Color palette to apply. One of `blue`, `custom_bg`, `custom_image`, `custom_text`, `gray_on_white`, `grey`, `green`, `orange`, `red`, `red_on_white`, `white_on_gray`, `white_on_green`, `green_on_white`, `white_on_red`, `white_on_yellow`, `yellow_on_white`, `black_on_light_yellow`, `black_on_light_green` or `black_on_light_red`.
- **value** (Number, Required) Value for the comparator.

Optional:

- **custom_bg_color** (String, Optional) Color palette to apply to the background, same values available as palette.
- **custom_fg_color** (String, Optional) Color palette to apply to the foreground, same values available as palette.
- **hide_value** (Boolean, Optional) Setting this to True hides values.
- **image_url** (String, Optional) Displays an image as the background.
- **metric** (String, Optional) Metric from the request to correlate this conditional format with.
- **timeframe** (String, Optional) Defines the displayed timeframe.


<a id="nestedblock--widget--toplist_definition--request--log_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--toplist_definition--request--style--compute_query"></a>
### Nested Schema for `widget.toplist_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--toplist_definition--request--style--group_by"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--toplist_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--toplist_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.toplist_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--toplist_definition--request--process_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **metric** (String, Required) Your chosen metric.

Optional:

- **filter_by** (List of String, Optional) List of processes.
- **limit** (Number, Optional) Max number of items in the filter list.
- **search_by** (String, Optional) Your chosen search term.


<a id="nestedblock--widget--toplist_definition--request--rum_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--toplist_definition--request--style--compute_query"></a>
### Nested Schema for `widget.toplist_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--toplist_definition--request--style--group_by"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--toplist_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--toplist_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.toplist_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--toplist_definition--request--security_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **index** (String, Required) Name of the index to query.

Optional:

- **compute** (Map of String, Optional, Deprecated) One of `compute` or `multi_compute` is required. The map has the keys as below. **Deprecated.** Define `compute_query` list with one element instead.
- **compute_query** (Block List, Max: 1) One of `compute_query` or `multi_compute` is required. The map has the keys as below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--compute_query))
- **group_by** (Block List) Multiple `group_by` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by))
- **multi_compute** (Block List) One of `compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--multi_compute))
- **search** (Map of String, Optional, Deprecated) Map defining the search query to use. **Deprecated.** Define `search_query` directly instead.
- **search_query** (String, Optional) The search query to use.

<a id="nestedblock--widget--toplist_definition--request--style--compute_query"></a>
### Nested Schema for `widget.toplist_definition.request.style.compute_query`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.


<a id="nestedblock--widget--toplist_definition--request--style--group_by"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by`

Optional:

- **facet** (String, Optional) Facet name.
- **limit** (Number, Optional) Maximum number of items in the group.
- **sort** (Map of String, Optional, Deprecated) One map is allowed with the keys as below. **Deprecated.** Define `sort_query` list with one element instead.
- **sort_query** (Block List, Max: 1) List of exactly one element describing the sort query to use. (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by--sort_query))

<a id="nestedblock--widget--toplist_definition--request--style--group_by--sort_query"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by.sort_query`

Required:

- **aggregation** (String, Required) The aggregation method.
- **order** (String, Required) Widget sorting methods.

Optional:

- **facet** (String, Optional) Facet name.



<a id="nestedblock--widget--toplist_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.toplist_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required) The aggregation method.

Optional:

- **facet** (String, Optional) Facet name.
- **interval** (Number, Optional) Define a time interval in seconds.



<a id="nestedblock--widget--toplist_definition--request--style"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Optional:

- **palette** (String, Optional) Color palette to apply to the widget. The available options are available here: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.




<a id="nestedblock--widget--trace_service_definition"></a>
### Nested Schema for `widget.trace_service_definition`

Required:

- **env** (String, Required) APM environment.
- **service** (String, Required) APM service.
- **span_name** (String, Required) APM span name

Optional:

- **display_format** (String, Optional) Number of columns to display. Available values are: `one_column`, `two_column`, or `three_column`.
- **live_span** (String, Optional) The timeframe to use when displaying the widget. One of `10m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `1y`, `alert`.
- **show_breakdown** (Boolean, Optional) Whether to show the latency breakdown or not.
- **show_distribution** (Boolean, Optional) Whether to show the latency distribution or not.
- **show_errors** (Boolean, Optional) Whether to show the error metrics or not.
- **show_hits** (Boolean, Optional) Whether to show the hits metrics or not
- **show_latency** (Boolean, Optional) Whether to show the latency metrics or not.
- **show_resource_list** (Boolean, Optional) Whether to show the resource list or not.
- **size_format** (String, Optional) Size of the widget. Available values are: `small`, `medium`, or `large`.
- **time** (Map of String, Optional, Deprecated) Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. **Deprecated.** Define `live_span` directly in the widget definition instead.
- **title** (String, Optional) The title of the widget.
- **title_align** (String, Optional) The alignment of the widget's title. One of `left`, `center`, or `right`.
- **title_size** (String, Optional) The size of the widget's title. Default is 16.


<a id="nestedblock--widget--widget_layout"></a>
### Nested Schema for `widget.widget_layout`

Required:

- **height** (Number, Required) The height of the widget.
- **width** (Number, Required) The width of the widget.
- **x** (Number, Required) The position of the widget on the x (horizontal) axis. Should be greater or equal to 0.
- **y** (Number, Required) The position of the widget on the y (vertical) axis. Should be greater or equal to 0.



<a id="nestedblock--template_variable"></a>
### Nested Schema for `template_variable`

Required:

- **name** (String, Required) The name of the variable.

Optional:

- **default** (String, Optional) The default value for the template variable on dashboard load.
- **prefix** (String, Optional) The tag prefix associated with the variable. Only tags with this prefix will appear in the variable dropdown.


<a id="nestedblock--template_variable_preset"></a>
### Nested Schema for `template_variable_preset`

Required:

- **name** (String, Required) The name of the preset.
- **template_variable** (Block List, Min: 1) The template variable names and assumed values under the given preset (see [below for nested schema](#nestedblock--template_variable_preset--template_variable))

<a id="nestedblock--template_variable_preset--template_variable"></a>
### Nested Schema for `template_variable_preset.template_variable`

Required:

- **name** (String, Required) The name of the template variable
- **value** (String, Required) The value that should be assumed by the template variable in this preset

## Import

Import is supported using the following syntax:

```shell
terraform import datadog_dashboard.my_service_dashboard sv7-gyh-kas
```
