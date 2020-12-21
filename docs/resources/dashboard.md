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
  title         = "Ordered Layout Dashboard"
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
    servicemap_definition {
      service = "master-db"
      filters = ["env:prod","datacenter:us1.prod.dog"]
      title = "env: prod, datacenter:us1.prod.dog, service: master-db"
      title_size = "16"
      title_align = "left"
    }
    layout = {
      height = 43
      width = 32
      x = 5
      y = 5
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
            aggregation = "avg"
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
            aggregation = "avg"
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
          background_color = "pink"
          font_size = "14"
          text_align = "center"
          show_tick = true
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

  template_variable_preset {
    name = "preset_1"
    template_variable {
      name = "var_1"
      value = "host.dc"
    }
    template_variable {
      name = "var_2"
      value = "my_service"
    }
  }
}

# Example Free Layout
resource "datadog_dashboard" "free_dashboard" {
  title         = "Free Layout Dashboard"
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
      display_format = "countsAndList"
      hide_zero_counts = true
      query = "type:metric"
      show_last_triggered = false
      sort = "status,asc"
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
      value = "host.dc"
    }
    template_variable {
      name = "var_2"
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

- **alert_graph_definition** (Block List, Max: 1) The definition for a Alert Graph widget (see [below for nested schema](#nestedblock--widget--alert_graph_definition))
- **alert_value_definition** (Block List, Max: 1) The definition for a Alert Value widget (see [below for nested schema](#nestedblock--widget--alert_value_definition))
- **change_definition** (Block List, Max: 1) The definition for a Change  widget (see [below for nested schema](#nestedblock--widget--change_definition))
- **check_status_definition** (Block List, Max: 1) The definition for a Check Status widget (see [below for nested schema](#nestedblock--widget--check_status_definition))
- **distribution_definition** (Block List, Max: 1) The definition for a Distribution widget (see [below for nested schema](#nestedblock--widget--distribution_definition))
- **event_stream_definition** (Block List, Max: 1) The definition for a Event Stream widget (see [below for nested schema](#nestedblock--widget--event_stream_definition))
- **event_timeline_definition** (Block List, Max: 1) The definition for a Event Timeline widget (see [below for nested schema](#nestedblock--widget--event_timeline_definition))
- **free_text_definition** (Block List, Max: 1) The definition for a Free Text widget (see [below for nested schema](#nestedblock--widget--free_text_definition))
- **group_definition** (Block List, Max: 1) The definition for a Group widget (see [below for nested schema](#nestedblock--widget--group_definition))
- **heatmap_definition** (Block List, Max: 1) The definition for a Heatmap widget (see [below for nested schema](#nestedblock--widget--heatmap_definition))
- **hostmap_definition** (Block List, Max: 1) The definition for a Hostmap widget (see [below for nested schema](#nestedblock--widget--hostmap_definition))
- **iframe_definition** (Block List, Max: 1) The definition for an Iframe widget (see [below for nested schema](#nestedblock--widget--iframe_definition))
- **image_definition** (Block List, Max: 1) The definition for an Image widget (see [below for nested schema](#nestedblock--widget--image_definition))
- **layout** (Map of String, Optional) The layout of the widget on a 'free' dashboard
- **log_stream_definition** (Block List, Max: 1) The definition for an Log Stream widget (see [below for nested schema](#nestedblock--widget--log_stream_definition))
- **manage_status_definition** (Block List, Max: 1) The definition for an Manage Status widget (see [below for nested schema](#nestedblock--widget--manage_status_definition))
- **note_definition** (Block List, Max: 1) The definition for a Note widget (see [below for nested schema](#nestedblock--widget--note_definition))
- **query_table_definition** (Block List, Max: 1) The definition for a Query Table widget (see [below for nested schema](#nestedblock--widget--query_table_definition))
- **query_value_definition** (Block List, Max: 1) The definition for a Query Value widget (see [below for nested schema](#nestedblock--widget--query_value_definition))
- **scatterplot_definition** (Block List, Max: 1) The definition for a Scatterplot widget (see [below for nested schema](#nestedblock--widget--scatterplot_definition))
- **service_level_objective_definition** (Block List, Max: 1) The definition for a Service Level Objective widget (see [below for nested schema](#nestedblock--widget--service_level_objective_definition))
- **servicemap_definition** (Block List, Max: 1) The definition for a Service Map widget (see [below for nested schema](#nestedblock--widget--servicemap_definition))
- **timeseries_definition** (Block List, Max: 1) The definition for a Timeseries widget (see [below for nested schema](#nestedblock--widget--timeseries_definition))
- **toplist_definition** (Block List, Max: 1) The definition for a Toplist widget (see [below for nested schema](#nestedblock--widget--toplist_definition))
- **trace_service_definition** (Block List, Max: 1) The definition for a Trace Service widget (see [below for nested schema](#nestedblock--widget--trace_service_definition))

<a id="nestedblock--widget--alert_graph_definition"></a>
### Nested Schema for `widget.alert_graph_definition`

Required:

- **alert_id** (String, Required)
- **viz_type** (String, Required)

Optional:

- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--alert_value_definition"></a>
### Nested Schema for `widget.alert_value_definition`

Required:

- **alert_id** (String, Required)

Optional:

- **precision** (Number, Optional)
- **text_align** (String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)
- **unit** (String, Optional)


<a id="nestedblock--widget--change_definition"></a>
### Nested Schema for `widget.change_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--custom_link))
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--request))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--change_definition--custom_link"></a>
### Nested Schema for `widget.change_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--change_definition--request"></a>
### Nested Schema for `widget.change_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--change_definition--request--apm_query))
- **change_type** (String, Optional)
- **compare_to** (String, Optional)
- **increase_good** (Boolean, Optional)
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--change_definition--request--log_query))
- **order_by** (String, Optional)
- **order_dir** (String, Optional)
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--change_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--change_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--change_definition--request--security_query))
- **show_present** (Boolean, Optional)

<a id="nestedblock--widget--change_definition--request--apm_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--change_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--change_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.change_definition.request.show_present.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--change_definition--request--log_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--change_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--change_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.change_definition.request.show_present.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--change_definition--request--process_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--change_definition--request--rum_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--change_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--change_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.change_definition.request.show_present.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--change_definition--request--security_query"></a>
### Nested Schema for `widget.change_definition.request.show_present`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--change_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--change_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.change_definition.request.show_present.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--change_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.change_definition.request.show_present.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--check_status_definition"></a>
### Nested Schema for `widget.check_status_definition`

Required:

- **check** (String, Required)
- **grouping** (String, Required)

Optional:

- **group** (String, Optional)
- **group_by** (List of String, Optional)
- **tags** (List of String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--distribution_definition"></a>
### Nested Schema for `widget.distribution_definition`

Optional:

- **legend_size** (String, Optional)
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--distribution_definition--request))
- **show_legend** (Boolean, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--distribution_definition--request"></a>
### Nested Schema for `widget.distribution_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--security_query))
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style))

<a id="nestedblock--widget--distribution_definition--request--apm_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--distribution_definition--request--style--group_by"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--distribution_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.distribution_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--distribution_definition--request--log_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--distribution_definition--request--style--group_by"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--distribution_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.distribution_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--distribution_definition--request--process_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--distribution_definition--request--rum_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--distribution_definition--request--style--group_by"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--distribution_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.distribution_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--distribution_definition--request--security_query"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--distribution_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--distribution_definition--request--style--group_by"></a>
### Nested Schema for `widget.distribution_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--distribution_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.distribution_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--distribution_definition--request--style"></a>
### Nested Schema for `widget.distribution_definition.request.style`

Optional:

- **palette** (String, Optional)




<a id="nestedblock--widget--event_stream_definition"></a>
### Nested Schema for `widget.event_stream_definition`

Required:

- **query** (String, Required)

Optional:

- **event_size** (String, Optional)
- **tags_execution** (String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--event_timeline_definition"></a>
### Nested Schema for `widget.event_timeline_definition`

Required:

- **query** (String, Required)

Optional:

- **tags_execution** (String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--free_text_definition"></a>
### Nested Schema for `widget.free_text_definition`

Required:

- **text** (String, Required)

Optional:

- **color** (String, Optional)
- **font_size** (String, Optional)
- **text_align** (String, Optional)


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

- **alert_graph_definition** (Block List, Max: 1) The definition for a Alert Graph widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--alert_graph_definition))
- **alert_value_definition** (Block List, Max: 1) The definition for a Alert Value widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--alert_value_definition))
- **change_definition** (Block List, Max: 1) The definition for a Change  widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--change_definition))
- **check_status_definition** (Block List, Max: 1) The definition for a Check Status widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--check_status_definition))
- **distribution_definition** (Block List, Max: 1) The definition for a Distribution widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--distribution_definition))
- **event_stream_definition** (Block List, Max: 1) The definition for a Event Stream widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--event_stream_definition))
- **event_timeline_definition** (Block List, Max: 1) The definition for a Event Timeline widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--event_timeline_definition))
- **free_text_definition** (Block List, Max: 1) The definition for a Free Text widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--free_text_definition))
- **heatmap_definition** (Block List, Max: 1) The definition for a Heatmap widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--heatmap_definition))
- **hostmap_definition** (Block List, Max: 1) The definition for a Hostmap widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--hostmap_definition))
- **iframe_definition** (Block List, Max: 1) The definition for an Iframe widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--iframe_definition))
- **image_definition** (Block List, Max: 1) The definition for an Image widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--image_definition))
- **layout** (Map of String, Optional) The layout of the widget on a 'free' dashboard
- **log_stream_definition** (Block List, Max: 1) The definition for an Log Stream widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--log_stream_definition))
- **manage_status_definition** (Block List, Max: 1) The definition for an Manage Status widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--manage_status_definition))
- **note_definition** (Block List, Max: 1) The definition for a Note widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--note_definition))
- **query_table_definition** (Block List, Max: 1) The definition for a Query Table widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--query_table_definition))
- **query_value_definition** (Block List, Max: 1) The definition for a Query Value widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--query_value_definition))
- **scatterplot_definition** (Block List, Max: 1) The definition for a Scatterplot widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--scatterplot_definition))
- **service_level_objective_definition** (Block List, Max: 1) The definition for a Service Level Objective widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--service_level_objective_definition))
- **servicemap_definition** (Block List, Max: 1) The definition for a Service Map widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--servicemap_definition))
- **timeseries_definition** (Block List, Max: 1) The definition for a Timeseries widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--timeseries_definition))
- **toplist_definition** (Block List, Max: 1) The definition for a Toplist widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--toplist_definition))
- **trace_service_definition** (Block List, Max: 1) The definition for a Trace Service widget (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition))

<a id="nestedblock--widget--group_definition--widget--alert_graph_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **alert_id** (String, Required)
- **viz_type** (String, Required)

Optional:

- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--alert_value_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **alert_id** (String, Required)

Optional:

- **precision** (Number, Optional)
- **text_align** (String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)
- **unit** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--change_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--custom_link))
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query))
- **change_type** (String, Optional)
- **compare_to** (String, Optional)
- **increase_good** (Boolean, Optional)
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query))
- **order_by** (String, Optional)
- **order_dir** (String, Optional)
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query))
- **show_present** (Boolean, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--show_present--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.show_present.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--group_definition--widget--check_status_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **check** (String, Required)
- **grouping** (String, Required)

Optional:

- **group** (String, Optional)
- **group_by** (List of String, Optional)
- **tags** (List of String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--distribution_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **legend_size** (String, Optional)
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request))
- **show_legend** (Boolean, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query))
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Optional:

- **palette** (String, Optional)




<a id="nestedblock--widget--group_definition--widget--event_stream_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **query** (String, Required)

Optional:

- **event_size** (String, Optional)
- **tags_execution** (String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--event_timeline_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **query** (String, Required)

Optional:

- **tags_execution** (String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--free_text_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **text** (String, Required)

Optional:

- **color** (String, Optional)
- **font_size** (String, Optional)
- **text_align** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--heatmap_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--custom_link))
- **event** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--event))
- **legend_size** (String, Optional)
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request))
- **show_legend** (Boolean, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)
- **yaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--yaxis))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--event"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.event`

Required:

- **q** (String, Required)

Optional:

- **tags_execution** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query))
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Optional:

- **palette** (String, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--yaxis"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.yaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)



<a id="nestedblock--widget--group_definition--widget--hostmap_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--custom_link))
- **group** (List of String, Optional)
- **no_group_hosts** (Boolean, Optional)
- **no_metric_hosts** (Boolean, Optional)
- **node_type** (String, Optional)
- **request** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request))
- **scope** (List of String, Optional)
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--style))
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request`

Optional:

- **fill** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--fill))
- **size** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--fill"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)




<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--group_definition--widget--trace_service_definition--style"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.style`

Optional:

- **fill_max** (String, Optional)
- **fill_min** (String, Optional)
- **palette** (String, Optional)
- **palette_flip** (Boolean, Optional)



<a id="nestedblock--widget--group_definition--widget--iframe_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **url** (String, Required)


<a id="nestedblock--widget--group_definition--widget--image_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **url** (String, Required)

Optional:

- **margin** (String, Optional)
- **sizing** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--log_stream_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **columns** (List of String, Optional)
- **indexes** (List of String, Optional)
- **logset** (String, Optional, Deprecated)
- **message_display** (String, Optional) One of: ['inline', 'expanded-md', 'expanded-lg']
- **query** (String, Optional)
- **show_date_column** (Boolean, Optional)
- **show_message_column** (Boolean, Optional)
- **sort** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--sort))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--sort"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.sort`

Required:

- **column** (String, Required)
- **order** (String, Required)



<a id="nestedblock--widget--group_definition--widget--manage_status_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **query** (String, Required)

Optional:

- **color_preference** (String, Optional)
- **count** (Number, Optional, Deprecated)
- **display_format** (String, Optional)
- **hide_zero_counts** (Boolean, Optional)
- **show_last_triggered** (Boolean, Optional)
- **sort** (String, Optional)
- **start** (Number, Optional, Deprecated)
- **summary_type** (String, Optional) One of: ['monitors', 'groups', 'combined']
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--note_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **content** (String, Required)

Optional:

- **background_color** (String, Optional)
- **font_size** (String, Optional)
- **show_tick** (Boolean, Optional)
- **text_align** (String, Optional)
- **tick_edge** (String, Optional)
- **tick_pos** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--query_table_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--custom_link))
- **has_search_bar** (String, Optional)
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request`

Optional:

- **aggregator** (String, Optional)
- **alias** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query))
- **apm_stats_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_stats_query))
- **cell_display_mode** (List of String, Optional)
- **conditional_formats** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--conditional_formats))
- **limit** (Number, Optional)
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query))
- **order** (String, Optional)
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_stats_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **env** (String, Required)
- **name** (String, Required)
- **primary_tag** (String, Required)
- **row_type** (String, Required)
- **service** (String, Required)

Optional:

- **columns** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--columns))
- **resource** (String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--columns"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.resource`

Required:

- **name** (String, Required)

Optional:

- **alias** (String, Optional)
- **cell_display_mode** (String, Optional)
- **order** (String, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--conditional_formats"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **comparator** (String, Required)
- **palette** (String, Required)
- **value** (Number, Required)

Optional:

- **custom_bg_color** (String, Optional)
- **custom_fg_color** (String, Optional)
- **hide_value** (Boolean, Optional)
- **image_url** (String, Optional)
- **metric** (String, Optional)
- **timeframe** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--group_definition--widget--query_value_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **autoscale** (Boolean, Optional)
- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--custom_link))
- **custom_unit** (String, Optional)
- **precision** (Number, Optional)
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request))
- **text_align** (String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request`

Optional:

- **aggregator** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query))
- **conditional_formats** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--conditional_formats))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--conditional_formats"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **comparator** (String, Required)
- **palette** (String, Required)
- **value** (Number, Required)

Optional:

- **custom_bg_color** (String, Optional)
- **custom_fg_color** (String, Optional)
- **hide_value** (Boolean, Optional)
- **image_url** (String, Optional)
- **metric** (String, Optional)
- **timeframe** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--group_definition--widget--scatterplot_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **color_by_groups** (List of String, Optional)
- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--custom_link))
- **request** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)
- **xaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--xaxis))
- **yaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--yaxis))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request`

Optional:

- **x** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--x))
- **y** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--x"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y`

Optional:

- **aggregator** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)




<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y`

Optional:

- **aggregator** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--group_definition--widget--trace_service_definition--xaxis"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.xaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--yaxis"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.yaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)



<a id="nestedblock--widget--group_definition--widget--service_level_objective_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **slo_id** (String, Required)
- **time_windows** (List of String, Required)
- **view_mode** (String, Required)
- **view_type** (String, Required)

Optional:

- **show_error_budget** (Boolean, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--servicemap_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **filters** (List of String, Required)
- **service** (String, Required)

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--custom_link))
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)



<a id="nestedblock--widget--group_definition--widget--timeseries_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--custom_link))
- **event** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--event))
- **legend_size** (String, Optional)
- **marker** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--marker))
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request))
- **right_yaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--right_yaxis))
- **show_legend** (Boolean, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)
- **yaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--yaxis))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--event"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.event`

Required:

- **q** (String, Required)

Optional:

- **tags_execution** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--marker"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.marker`

Required:

- **value** (String, Required)

Optional:

- **display_type** (String, Optional)
- **label** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query))
- **display_type** (String, Optional)
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query))
- **metadata** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--metadata))
- **network_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--network_query))
- **on_right_yaxis** (Boolean, Optional)
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query))
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--metadata"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **expression** (String, Required)

Optional:

- **alias_name** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--network_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Optional:

- **line_type** (String, Optional)
- **line_width** (String, Optional)
- **palette** (String, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--right_yaxis"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.right_yaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--yaxis"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.yaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)



<a id="nestedblock--widget--group_definition--widget--toplist_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--custom_link))
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--custom_link"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query))
- **conditional_formats** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--conditional_formats))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query))
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style))

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--apm_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--conditional_formats"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **comparator** (String, Required)
- **palette** (String, Required)
- **value** (Number, Required)

Optional:

- **custom_bg_color** (String, Optional)
- **custom_fg_color** (String, Optional)
- **hide_value** (Boolean, Optional)
- **image_url** (String, Optional)
- **metric** (String, Optional)
- **timeframe** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--log_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--process_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--rum_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--security_query"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--group_by"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--group_definition--widget--trace_service_definition--request--style"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition.request.style`

Optional:

- **palette** (String, Optional)




<a id="nestedblock--widget--group_definition--widget--trace_service_definition"></a>
### Nested Schema for `widget.group_definition.widget.trace_service_definition`

Required:

- **env** (String, Required)
- **service** (String, Required)
- **span_name** (String, Required)

Optional:

- **display_format** (String, Optional)
- **show_breakdown** (Boolean, Optional)
- **show_distribution** (Boolean, Optional)
- **show_errors** (Boolean, Optional)
- **show_hits** (Boolean, Optional)
- **show_latency** (Boolean, Optional)
- **show_resource_list** (Boolean, Optional)
- **size_format** (String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)




<a id="nestedblock--widget--heatmap_definition"></a>
### Nested Schema for `widget.heatmap_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--custom_link))
- **event** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--event))
- **legend_size** (String, Optional)
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request))
- **show_legend** (Boolean, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)
- **yaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--heatmap_definition--yaxis))

<a id="nestedblock--widget--heatmap_definition--custom_link"></a>
### Nested Schema for `widget.heatmap_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--heatmap_definition--event"></a>
### Nested Schema for `widget.heatmap_definition.event`

Required:

- **q** (String, Required)

Optional:

- **tags_execution** (String, Optional)


<a id="nestedblock--widget--heatmap_definition--request"></a>
### Nested Schema for `widget.heatmap_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--security_query))
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style))

<a id="nestedblock--widget--heatmap_definition--request--apm_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--heatmap_definition--request--style--group_by"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--heatmap_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.heatmap_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--heatmap_definition--request--log_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--heatmap_definition--request--style--group_by"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--heatmap_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.heatmap_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--heatmap_definition--request--process_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--heatmap_definition--request--rum_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--heatmap_definition--request--style--group_by"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--heatmap_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.heatmap_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--heatmap_definition--request--security_query"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--heatmap_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--heatmap_definition--request--style--group_by"></a>
### Nested Schema for `widget.heatmap_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--heatmap_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.heatmap_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--heatmap_definition--request--style"></a>
### Nested Schema for `widget.heatmap_definition.request.style`

Optional:

- **palette** (String, Optional)



<a id="nestedblock--widget--heatmap_definition--yaxis"></a>
### Nested Schema for `widget.heatmap_definition.yaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)



<a id="nestedblock--widget--hostmap_definition"></a>
### Nested Schema for `widget.hostmap_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--custom_link))
- **group** (List of String, Optional)
- **no_group_hosts** (Boolean, Optional)
- **no_metric_hosts** (Boolean, Optional)
- **node_type** (String, Optional)
- **request** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request))
- **scope** (List of String, Optional)
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--style))
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--hostmap_definition--custom_link"></a>
### Nested Schema for `widget.hostmap_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--hostmap_definition--request"></a>
### Nested Schema for `widget.hostmap_definition.request`

Optional:

- **fill** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--fill))
- **size** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size))

<a id="nestedblock--widget--hostmap_definition--request--fill"></a>
### Nested Schema for `widget.hostmap_definition.request.size`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query))

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--hostmap_definition--request--size--log_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--hostmap_definition--request--size--log_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--log_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--hostmap_definition--request--size--process_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.process_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--rum_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--hostmap_definition--request--size--security_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--hostmap_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)




<a id="nestedblock--widget--hostmap_definition--request--size"></a>
### Nested Schema for `widget.hostmap_definition.request.size`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query))

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--apm_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--apm_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.apm_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--hostmap_definition--request--size--log_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--log_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--hostmap_definition--request--size--log_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--log_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.log_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--hostmap_definition--request--size--process_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.process_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--rum_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--rum_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--rum_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.rum_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--hostmap_definition--request--size--security_query"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--hostmap_definition--request--size--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--hostmap_definition--request--size--security_query--group_by"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--hostmap_definition--request--size--security_query--multi_compute"></a>
### Nested Schema for `widget.hostmap_definition.request.size.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--hostmap_definition--style"></a>
### Nested Schema for `widget.hostmap_definition.style`

Optional:

- **fill_max** (String, Optional)
- **fill_min** (String, Optional)
- **palette** (String, Optional)
- **palette_flip** (Boolean, Optional)



<a id="nestedblock--widget--iframe_definition"></a>
### Nested Schema for `widget.iframe_definition`

Required:

- **url** (String, Required)


<a id="nestedblock--widget--image_definition"></a>
### Nested Schema for `widget.image_definition`

Required:

- **url** (String, Required)

Optional:

- **margin** (String, Optional)
- **sizing** (String, Optional)


<a id="nestedblock--widget--log_stream_definition"></a>
### Nested Schema for `widget.log_stream_definition`

Optional:

- **columns** (List of String, Optional)
- **indexes** (List of String, Optional)
- **logset** (String, Optional, Deprecated)
- **message_display** (String, Optional) One of: ['inline', 'expanded-md', 'expanded-lg']
- **query** (String, Optional)
- **show_date_column** (Boolean, Optional)
- **show_message_column** (Boolean, Optional)
- **sort** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--log_stream_definition--sort))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--log_stream_definition--sort"></a>
### Nested Schema for `widget.log_stream_definition.sort`

Required:

- **column** (String, Required)
- **order** (String, Required)



<a id="nestedblock--widget--manage_status_definition"></a>
### Nested Schema for `widget.manage_status_definition`

Required:

- **query** (String, Required)

Optional:

- **color_preference** (String, Optional)
- **count** (Number, Optional, Deprecated)
- **display_format** (String, Optional)
- **hide_zero_counts** (Boolean, Optional)
- **show_last_triggered** (Boolean, Optional)
- **sort** (String, Optional)
- **start** (Number, Optional, Deprecated)
- **summary_type** (String, Optional) One of: ['monitors', 'groups', 'combined']
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--note_definition"></a>
### Nested Schema for `widget.note_definition`

Required:

- **content** (String, Required)

Optional:

- **background_color** (String, Optional)
- **font_size** (String, Optional)
- **show_tick** (Boolean, Optional)
- **text_align** (String, Optional)
- **tick_edge** (String, Optional)
- **tick_pos** (String, Optional)


<a id="nestedblock--widget--query_table_definition"></a>
### Nested Schema for `widget.query_table_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--custom_link))
- **has_search_bar** (String, Optional)
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--query_table_definition--custom_link"></a>
### Nested Schema for `widget.query_table_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--query_table_definition--request"></a>
### Nested Schema for `widget.query_table_definition.request`

Optional:

- **aggregator** (String, Optional)
- **alias** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--apm_query))
- **apm_stats_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--apm_stats_query))
- **cell_display_mode** (List of String, Optional)
- **conditional_formats** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--conditional_formats))
- **limit** (Number, Optional)
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--log_query))
- **order** (String, Optional)
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query))

<a id="nestedblock--widget--query_table_definition--request--apm_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--query_table_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--query_table_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--query_table_definition--request--apm_stats_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **env** (String, Required)
- **name** (String, Required)
- **primary_tag** (String, Required)
- **row_type** (String, Required)
- **service** (String, Required)

Optional:

- **columns** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--columns))
- **resource** (String, Optional)

<a id="nestedblock--widget--query_table_definition--request--security_query--columns"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.columns`

Required:

- **name** (String, Required)

Optional:

- **alias** (String, Optional)
- **cell_display_mode** (String, Optional)
- **order** (String, Optional)



<a id="nestedblock--widget--query_table_definition--request--conditional_formats"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **comparator** (String, Required)
- **palette** (String, Required)
- **value** (Number, Required)

Optional:

- **custom_bg_color** (String, Optional)
- **custom_fg_color** (String, Optional)
- **hide_value** (Boolean, Optional)
- **image_url** (String, Optional)
- **metric** (String, Optional)
- **timeframe** (String, Optional)


<a id="nestedblock--widget--query_table_definition--request--log_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--query_table_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--query_table_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--query_table_definition--request--process_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--query_table_definition--request--rum_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--query_table_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--query_table_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--query_table_definition--request--security_query"></a>
### Nested Schema for `widget.query_table_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--query_table_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--query_table_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--query_table_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_table_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--query_value_definition"></a>
### Nested Schema for `widget.query_value_definition`

Optional:

- **autoscale** (Boolean, Optional)
- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--custom_link))
- **custom_unit** (String, Optional)
- **precision** (Number, Optional)
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request))
- **text_align** (String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--query_value_definition--custom_link"></a>
### Nested Schema for `widget.query_value_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--query_value_definition--request"></a>
### Nested Schema for `widget.query_value_definition.request`

Optional:

- **aggregator** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--apm_query))
- **conditional_formats** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--conditional_formats))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query))

<a id="nestedblock--widget--query_value_definition--request--apm_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--query_value_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--query_value_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--query_value_definition--request--conditional_formats"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **comparator** (String, Required)
- **palette** (String, Required)
- **value** (Number, Required)

Optional:

- **custom_bg_color** (String, Optional)
- **custom_fg_color** (String, Optional)
- **hide_value** (Boolean, Optional)
- **image_url** (String, Optional)
- **metric** (String, Optional)
- **timeframe** (String, Optional)


<a id="nestedblock--widget--query_value_definition--request--log_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--query_value_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--query_value_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--query_value_definition--request--process_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--query_value_definition--request--rum_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--query_value_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--query_value_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--query_value_definition--request--security_query"></a>
### Nested Schema for `widget.query_value_definition.request.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--query_value_definition--request--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--query_value_definition--request--security_query--group_by"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--query_value_definition--request--security_query--multi_compute"></a>
### Nested Schema for `widget.query_value_definition.request.security_query.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--scatterplot_definition"></a>
### Nested Schema for `widget.scatterplot_definition`

Optional:

- **color_by_groups** (List of String, Optional)
- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--custom_link))
- **request** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)
- **xaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--xaxis))
- **yaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--yaxis))

<a id="nestedblock--widget--scatterplot_definition--custom_link"></a>
### Nested Schema for `widget.scatterplot_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--scatterplot_definition--request"></a>
### Nested Schema for `widget.scatterplot_definition.request`

Optional:

- **x** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--x))
- **y** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y))

<a id="nestedblock--widget--scatterplot_definition--request--x"></a>
### Nested Schema for `widget.scatterplot_definition.request.y`

Optional:

- **aggregator** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--scatterplot_definition--request--y--log_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--scatterplot_definition--request--y--process_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.process_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--scatterplot_definition--request--y--security_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)




<a id="nestedblock--widget--scatterplot_definition--request--y"></a>
### Nested Schema for `widget.scatterplot_definition.request.y`

Optional:

- **aggregator** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query))

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--apm_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--apm_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.apm_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--scatterplot_definition--request--y--log_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--log_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--log_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.log_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--scatterplot_definition--request--y--process_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.process_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--rum_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--rum_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.rum_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--scatterplot_definition--request--y--security_query"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--scatterplot_definition--request--y--security_query--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--group_by"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--scatterplot_definition--request--y--security_query--multi_compute"></a>
### Nested Schema for `widget.scatterplot_definition.request.y.security_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)





<a id="nestedblock--widget--scatterplot_definition--xaxis"></a>
### Nested Schema for `widget.scatterplot_definition.xaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)


<a id="nestedblock--widget--scatterplot_definition--yaxis"></a>
### Nested Schema for `widget.scatterplot_definition.yaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)



<a id="nestedblock--widget--service_level_objective_definition"></a>
### Nested Schema for `widget.service_level_objective_definition`

Required:

- **slo_id** (String, Required)
- **time_windows** (List of String, Required)
- **view_mode** (String, Required)
- **view_type** (String, Required)

Optional:

- **show_error_budget** (Boolean, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)


<a id="nestedblock--widget--servicemap_definition"></a>
### Nested Schema for `widget.servicemap_definition`

Required:

- **filters** (List of String, Required)
- **service** (String, Required)

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--servicemap_definition--custom_link))
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--servicemap_definition--custom_link"></a>
### Nested Schema for `widget.servicemap_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)



<a id="nestedblock--widget--timeseries_definition"></a>
### Nested Schema for `widget.timeseries_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--custom_link))
- **event** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--event))
- **legend_size** (String, Optional)
- **marker** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--marker))
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request))
- **right_yaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--timeseries_definition--right_yaxis))
- **show_legend** (Boolean, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)
- **yaxis** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--timeseries_definition--yaxis))

<a id="nestedblock--widget--timeseries_definition--custom_link"></a>
### Nested Schema for `widget.timeseries_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--timeseries_definition--event"></a>
### Nested Schema for `widget.timeseries_definition.event`

Required:

- **q** (String, Required)

Optional:

- **tags_execution** (String, Optional)


<a id="nestedblock--widget--timeseries_definition--marker"></a>
### Nested Schema for `widget.timeseries_definition.marker`

Required:

- **value** (String, Required)

Optional:

- **display_type** (String, Optional)
- **label** (String, Optional)


<a id="nestedblock--widget--timeseries_definition--request"></a>
### Nested Schema for `widget.timeseries_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--apm_query))
- **display_type** (String, Optional)
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--log_query))
- **metadata** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--metadata))
- **network_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--network_query))
- **on_right_yaxis** (Boolean, Optional)
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--security_query))
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style))

<a id="nestedblock--widget--timeseries_definition--request--apm_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--timeseries_definition--request--log_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--timeseries_definition--request--metadata"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **expression** (String, Required)

Optional:

- **alias_name** (String, Optional)


<a id="nestedblock--widget--timeseries_definition--request--network_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--timeseries_definition--request--process_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--timeseries_definition--request--rum_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--timeseries_definition--request--security_query"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--timeseries_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--timeseries_definition--request--style--group_by"></a>
### Nested Schema for `widget.timeseries_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--timeseries_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.timeseries_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--timeseries_definition--request--style"></a>
### Nested Schema for `widget.timeseries_definition.request.style`

Optional:

- **line_type** (String, Optional)
- **line_width** (String, Optional)
- **palette** (String, Optional)



<a id="nestedblock--widget--timeseries_definition--right_yaxis"></a>
### Nested Schema for `widget.timeseries_definition.right_yaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)


<a id="nestedblock--widget--timeseries_definition--yaxis"></a>
### Nested Schema for `widget.timeseries_definition.yaxis`

Optional:

- **include_zero** (Boolean, Optional)
- **label** (String, Optional)
- **max** (String, Optional)
- **min** (String, Optional)
- **scale** (String, Optional)



<a id="nestedblock--widget--toplist_definition"></a>
### Nested Schema for `widget.toplist_definition`

Optional:

- **custom_link** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--custom_link))
- **request** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request))
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)

<a id="nestedblock--widget--toplist_definition--custom_link"></a>
### Nested Schema for `widget.toplist_definition.custom_link`

Required:

- **label** (String, Required)
- **link** (String, Required)


<a id="nestedblock--widget--toplist_definition--request"></a>
### Nested Schema for `widget.toplist_definition.request`

Optional:

- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--apm_query))
- **conditional_formats** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--conditional_formats))
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--log_query))
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--process_query))
- **q** (String, Optional)
- **rum_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--rum_query))
- **security_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--security_query))
- **style** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style))

<a id="nestedblock--widget--toplist_definition--request--apm_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--toplist_definition--request--style--group_by"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--toplist_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.toplist_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--toplist_definition--request--conditional_formats"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **comparator** (String, Required)
- **palette** (String, Required)
- **value** (Number, Required)

Optional:

- **custom_bg_color** (String, Optional)
- **custom_fg_color** (String, Optional)
- **hide_value** (Boolean, Optional)
- **image_url** (String, Optional)
- **metric** (String, Optional)
- **timeframe** (String, Optional)


<a id="nestedblock--widget--toplist_definition--request--log_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--toplist_definition--request--style--group_by"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--toplist_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.toplist_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--toplist_definition--request--process_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)


<a id="nestedblock--widget--toplist_definition--request--rum_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--toplist_definition--request--style--group_by"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--toplist_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.toplist_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--toplist_definition--request--security_query"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Required:

- **index** (String, Required)

Optional:

- **compute** (Map of String, Optional)
- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--group_by))
- **multi_compute** (Block List) (see [below for nested schema](#nestedblock--widget--toplist_definition--request--style--multi_compute))
- **search** (Map of String, Optional)

<a id="nestedblock--widget--toplist_definition--request--style--group_by"></a>
### Nested Schema for `widget.toplist_definition.request.style.group_by`

Optional:

- **facet** (String, Optional)
- **limit** (Number, Optional)
- **sort** (Map of String, Optional)


<a id="nestedblock--widget--toplist_definition--request--style--multi_compute"></a>
### Nested Schema for `widget.toplist_definition.request.style.multi_compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--widget--toplist_definition--request--style"></a>
### Nested Schema for `widget.toplist_definition.request.style`

Optional:

- **palette** (String, Optional)




<a id="nestedblock--widget--trace_service_definition"></a>
### Nested Schema for `widget.trace_service_definition`

Required:

- **env** (String, Required)
- **service** (String, Required)
- **span_name** (String, Required)

Optional:

- **display_format** (String, Optional)
- **show_breakdown** (Boolean, Optional)
- **show_distribution** (Boolean, Optional)
- **show_errors** (Boolean, Optional)
- **show_hits** (Boolean, Optional)
- **show_latency** (Boolean, Optional)
- **show_resource_list** (Boolean, Optional)
- **size_format** (String, Optional)
- **time** (Map of String, Optional)
- **title** (String, Optional)
- **title_align** (String, Optional)
- **title_size** (String, Optional)



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
