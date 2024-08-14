# Manage Datadog Powerpacks
resource "datadog_powerpack" "foo" {
  description = "Created using the Datadog provider in terraform"
  live_span   = "4h"

  layout {
    height = 10
    width  = 3
    x      = 1
    y      = 0
  }

  template_variables {
    defaults = ["defaults"]
    name     = "datacenter"
  }

  widget {
    event_stream_definition {
      query       = "*"
      event_size  = "l"
      title       = "Widget Title"
      title_size  = 16
      title_align = "right"
    }
  }
}