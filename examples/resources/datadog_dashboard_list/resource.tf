# Create a new Dashboard List with two Dashboards
resource "datadog_dashboard_list" "new_list" {
  depends_on = [
    datadog_dashboard.screen,
    datadog_dashboard.time
  ]

  name = "Terraform Created List"
  dash_item {
    type    = "custom_timeboard"
    dash_id = datadog_dashboard.time.id
  }
  dash_item {
    type    = "custom_screenboard"
    dash_id = datadog_dashboard.screen.id
  }
}

resource "datadog_dashboard" "time" {
  title        = "TF Test Layout Dashboard"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "ordered"
  is_read_only = true
  widget {
    alert_graph_definition {
      alert_id  = "1234"
      viz_type  = "timeseries"
      title     = "Widget Title"
      live_span = "1h"
    }
  }

}

resource "datadog_dashboard" "screen" {
  title        = "TF Test Free Layout Dashboard"
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
}
