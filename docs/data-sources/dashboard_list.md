---
page_title: "datadog_dashboard_list Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve information about an existing dashboard list, for use in other resources. In particular, it can be used in a dashboard to register it in the list.
---

# Data Source `datadog_dashboard_list`

Use this data source to retrieve information about an existing dashboard list, for use in other resources. In particular, it can be used in a dashboard to register it in the list.

## Example Usage

```terraform
data "datadog_dashboard_list" "test" {
  name = "My super list"
}
```

## Schema

### Required

- **name** (String) A dashboard list name to limit the search.

### Optional

- **id** (String) The ID of this resource.


