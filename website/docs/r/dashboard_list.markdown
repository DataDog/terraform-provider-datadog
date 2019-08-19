---
layout: "datadog"
page_title: "Datadog: datadog_dashboard_list"
sidebar_current: "docs-datadog-resource-dashboard-list"
description: |-
  Provides a Datadog dashboard list resource. This can be used to create and manage dashboard lists and their sub elements.
---

# datadog_dashboard_list

Provides a Datadog dashboard_list resource. This can be used to create and manage Datadog Dashboard Lists and the individual dashboards within them.

## Example Usage

Create a new Dashboard list with two dashbaords

```hcl
resource "datadog_dashboard_list" "new_list" {
    name = "Terraform Created List"
    dash_item {
        type = "custom_timeboard"
        dash_id = "<XYZ>"
    }
    dash_item {
        type = "custom_screenboard"
        dash_id = "<ABC>"
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of this Dashbaord List.
* `dash_item` - (Optional) An individual dashboard object to add to this Dashboard List. If present, must contain the following:
  * `type` - (Required) The type of this dashboard. Available options are: `custom_timeboard`, `custom_screenboard`, `integration_screenboard`, `integration_timeboard`, and `host_timeboard`
  * `dash_id` - (Required) The ID of this dashboard.


## Import

dashboard lists can be imported using their id, e.g.

```
$ terraform import datadog_dashboard_list.new_list 123456
```
