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
