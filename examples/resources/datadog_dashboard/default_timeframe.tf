resource "datadog_dashboard" "default_timeframe_dashboard" {
  title       = "Dashboard with default timeframe"
  layout_type = "ordered"

  default_timeframe {
    type  = "live"
    unit  = "week"
    value = 1
  }

  widget {
    note_definition {
      content = "Widget 1"
    }
  }
}
