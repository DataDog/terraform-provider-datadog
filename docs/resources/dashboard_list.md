---
page_title: "datadog_dashboard_list Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog dashboard_list resource. This can be used to create and manage Datadog Dashboard Lists and the individual dashboards within them.
---

# Resource `datadog_dashboard_list`

Provides a Datadog dashboard_list resource. This can be used to create and manage Datadog Dashboard Lists and the individual dashboards within them.

## Example Usage

```terraform
# Create a new Dashboard List with two Dashboards
resource "datadog_dashboard_list" "new_list" {
  depends_on = [
    "datadog_dashboard.screen",
    "datadog_dashboard.time"
  ]

  name = "Terraform Created List"
  dash_item {
    type    = "custom_timeboard"
    dash_id = "${datadog_dashboard.time.id}"
  }
  dash_item {
    type    = "custom_screenboard"
    dash_id = "${datadog_dashboard.screen.id}"
  }
}

resource "datadog_dashboard" "time" {
  title        = "TF Test Layout Dashboard"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "ordered"
  is_read_only = true
  widget {
    alert_graph_definition {
      alert_id = "1234"
      viz_type = "timeseries"
      title    = "Widget Title"
      time = {
        live_span = "1h"
      }
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
    layout = {
      height = 43
      width  = 32
      x      = 5
      y      = 5
    }
  }
}
```

## Schema

### Required

- **name** (String, Required) The name of the Dashboard List

### Optional

- **dash_item** (Block Set) A set of dashboard items that belong to this list (see [below for nested schema](#nestedblock--dash_item))
- **id** (String, Optional) The ID of this resource.

<a id="nestedblock--dash_item"></a>
### Nested Schema for `dash_item`

Required:

- **dash_id** (String, Required) The ID of the dashboard to add
- **type** (String, Required) The type of this dashboard. Available options are: `custom_timeboard`, `custom_screenboard`, `integration_screenboard`, `integration_timeboard`, and `host_timeboard`

## Import

Import is supported using the following syntax:

```shell
terraform import datadog_dashboard_list.new_list 123456
```
