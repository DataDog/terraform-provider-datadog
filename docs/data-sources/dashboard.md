---
page_title: "datadog_dashboard Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve information about an existing dashboard, for use in other resources. In particular, it can be used in a monitor message to link to a specific dashboard.
---

# Data Source `datadog_dashboard`

Use this data source to retrieve information about an existing dashboard, for use in other resources. In particular, it can be used in a monitor message to link to a specific dashboard.

## Example Usage

```terraform
data "datadog_dashboard" "test" {
  name = "My super dashboard"
}
```

## Schema

### Required

- **name** (String, Required) The dashboard name to search for. Must only match one dashboard.

### Optional

- **id** (String, Optional) The ID of this resource.

### Read-only

- **title** (String, Read-only) The name of the dashboard.
- **url** (String, Read-only) The URL to a specific dashboard.


