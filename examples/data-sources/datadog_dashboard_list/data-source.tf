data "datadog_dashboard_list" "test" {
  name = "My super list"
}

# Create a dashboard and register it in the list above.
resource "datadog_dashboard" "time" {
  title           = "TF Test Layout Dashboard"
  description     = "Created using the Datadog provider in Terraform"
  dashboard_lists = ["${data.datadog_dashboard_list.test.id}"]
  layout_type     = "ordered"
  is_read_only    = true
  widget {
    alert_graph_definition {
      alert_id  = "1234"
      viz_type  = "timeseries"
      title     = "Widget Title"
      live_span = "1h"
    }
  }
}
