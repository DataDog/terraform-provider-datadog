---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_teams Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve information about existing teams for use in other resources.
---

# datadog_teams (Data Source)

Use this data source to retrieve information about existing teams for use in other resources.

## Example Usage

```terraform
data "datadog_teams" "example" {
  filter_keyword = "team-member@company.com"
  filter_me      = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `filter_keyword` (String) Search query. Can be team name, team handle, or email of team member.
- `filter_me` (Boolean) When true, only returns teams the current user belongs to.

### Read-Only

- `id` (String) The ID of this resource.
- `teams` (List of Object) List of teams (see [below for nested schema](#nestedatt--teams))

<a id="nestedatt--teams"></a>
### Nested Schema for `teams`

Read-Only:

- `description` (String)
- `handle` (String)
- `id` (String)
- `link_count` (Number)
- `name` (String)
- `summary` (String)
- `user_count` (Number)