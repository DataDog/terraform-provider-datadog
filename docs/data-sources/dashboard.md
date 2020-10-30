---
page_title: "datadog_dashboard"
---

# datadog_dashboard Data Source

Use this data source to retrieve information about an existing dashboard, for use in other resources. In particular, it can be used in a monitor message to link to a specific dashboard.

## Example Usage

```
data "datadog_dashboard" "test" {
  name = "My super dashboard"
}
```

## Argument Reference

-   `name`: (Required) The dashboard name to search for. Must only match one dashboard.

## Attributes Reference

-   `id`: ID of the Datadog dashboard.
-   `title`: The name of the dashboard.
-   `url`: The URL to a specific dashboard.
