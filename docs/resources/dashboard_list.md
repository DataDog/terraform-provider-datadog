---
page_title: "datadog_dashboard_list"
---

# datadog_dashboard_list Resource

Provides a Datadog dashboard_list resource. This can be used to create and manage Datadog Dashboard Lists and the individual dashboards within them.

## Example Usage

Create a new Dashboard list with two dashboards

```hcl
resource "datadog_dashboard_list" "new_list" {
	depends_on = [
		"datadog_dashboard.screen",
		"datadog_dashboard.time"
	]

    name = "Terraform Created List"
    dash_item {
        type = "custom_timeboard"
        dash_id = "${datadog_dashboard.time.id}"
    }
    dash_item {
        type = "custom_screenboard"
        dash_id = "${datadog_dashboard.screen.id}"
	}
}

resource "datadog_dashboard" "time" {
	title         = "TF Test Layout Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = true
	widget {
		alert_graph_definition {
		  alert_id = "1234"
		  viz_type = "timeseries"
		  title = "Widget Title"
		  time = {
			live_span = "1h"
		  }
		}
	  }

}

resource "datadog_dashboard" "screen" {
	title         = "TF Test Free Layout Dashboard"
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

}
```

## Argument Reference

The following arguments are supported:

* `name`: (Required) The name of this Dashboard List.
* `dash_item`: (Optional) An individual dashboard object to add to this Dashboard List. If present, must contain the following:
  * `type`: (Required) The type of this dashboard. Available options are: `custom_timeboard`, `custom_screenboard`, `integration_screenboard`, `integration_timeboard`, and `host_timeboard`
  * `dash_id`: (Required) The ID of this dashboard.


## Import

dashboard lists can be imported using their id, e.g.

```
$ terraform import datadog_dashboard_list.new_list 123456
```
