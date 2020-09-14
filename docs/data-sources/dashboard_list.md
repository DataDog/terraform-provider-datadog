---
page_title: "datadog_dashboard_list"
---

# datadog_dashboard_list Data Source

Use this data source to retrieve information about an existing dashboard list, for use in other resources. In particular, it can be used in a dashboard to register it in the list.

## Example Usage

```
data "datadog_dashboard_list" "test" {
  name = "My super list"
}
```

## Argument Reference

- `name`: (Required) A dashboard list name to limit the search.

## Attributes Reference

- `id`: ID of the Datadog dashboard list.
