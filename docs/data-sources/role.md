---
page_title: "datadog_role"
---

# datadog_role Data Source

Use this data source to retrieve information about an existing role for use in other resources.

## Example Usage

```
data "datadog_role" "test" {
  filter = "Datadog Standard Role"
}
```

## Argument Reference

-   `filter`: (Required) A string on which to filter the roles.

~> **NOTE:** If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single role.

## Attributes Reference

-   `name`: Name of the role.
-   `user_count`: Number of users assigned to this role.
