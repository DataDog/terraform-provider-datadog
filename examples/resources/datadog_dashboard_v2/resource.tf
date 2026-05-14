# Example Ordered Layout
resource "datadog_dashboard_v2" "ordered_dashboard" {
  title       = "Ordered Layout Dashboard"
  description = "Created using the Datadog provider in Terraform"
  layout_type = "ordered"

  widget {
    timeseries_definition {
      request {
        formula {
          formula_expression = "my_query_1"
          alias              = "CPU Usage"
        }
        query {
          metric_query {
            data_source = "metrics"
            query       = "avg:system.cpu.user{*} by {env}"
            name        = "my_query_1"
            aggregator  = "avg"
          }
        }
      }
      title     = "CPU Usage by Environment"
      live_span = "1h"
    }
  }

  widget {
    query_value_definition {
      request {
        formula {
          formula_expression = "my_query_1"
        }
        query {
          metric_query {
            data_source = "metrics"
            query       = "avg:system.load.1{*}"
            name        = "my_query_1"
            aggregator  = "avg"
          }
        }
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
      autoscale  = true
      precision  = "4"
      text_align = "right"
      title      = "System Load"
      live_span  = "1h"
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
    }
  }

  template_variable {
    name    = "var_1"
    prefix  = "host"
    default = "aws"
  }
}

# Example Free Layout
resource "datadog_dashboard_v2" "free_dashboard" {
  title       = "Free Layout Dashboard"
  description = "Created using the Datadog provider in Terraform"
  layout_type = "free"

  widget {
    timeseries_definition {
      request {
        formula {
          formula_expression = "my_query_1"
        }
        query {
          metric_query {
            data_source = "metrics"
            query       = "avg:system.cpu.user{*} by {env}"
            name        = "my_query_1"
            aggregator  = "sum"
          }
        }
      }
    }
    widget_layout {
      height = 16
      width  = 25
      x      = 0
      y      = 0
    }
  }

  template_variable {
    name    = "var_1"
    prefix  = "host"
    default = "aws"
  }
}
