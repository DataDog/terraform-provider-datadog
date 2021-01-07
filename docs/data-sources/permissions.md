---
page_title: "datadog_role"
---

# datadog_permissions Data Source

Use this data source to retrieve the list of Datadog permissions by name and their corresponding ID, for use in the role resource.

## Example Usage

```
data "datadog_permissions" "permissions" {}
```

## Attributes Reference

-   `permissions`: Map of permissions names to their corresponding ID.
