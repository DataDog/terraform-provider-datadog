---
page_title: "datadog_security_monitoring_rules Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve information about existing security monitoring rules for use in other resources.
---

# Data Source `datadog_security_monitoring_rules`

Use this data source to retrieve information about existing security monitoring rules for use in other resources.

## Example Usage

```terraform
data "datadog_security_monitoring_rules" "test" {
  name_filter = "attack"
  tags_filter = ["foo:bar"]
  default_only_filter = true
}
```

## Schema

### Optional

- **default_only_filter** (Boolean) Limit the search to default rules
- **id** (String) The ID of this resource.
- **name_filter** (String) A rule name to limit the search
- **tags_filter** (List of String) A list of tags to limit the search
- **user_only_filter** (Boolean) Limit the search to user rules

### Read-only

- **rule_ids** (List of String) List of IDs of the matched rules.
- **rules** (List of Object) List of rules. (see [below for nested schema](#nestedatt--rules))

<a id="nestedatt--rules"></a>
### Nested Schema for `rules`

Read-only:

- **case** (List of Object) (see [below for nested schema](#nestedobjatt--rules--case))
- **enabled** (Boolean)
- **message** (String)
- **name** (String)
- **options** (List of Object) (see [below for nested schema](#nestedobjatt--rules--options))
- **query** (List of Object) (see [below for nested schema](#nestedobjatt--rules--query))
- **tags** (List of String)

<a id="nestedobjatt--rules--case"></a>
### Nested Schema for `rules.case`

Read-only:

- **condition** (String)
- **name** (String)
- **notifications** (List of String)
- **status** (String)


<a id="nestedobjatt--rules--options"></a>
### Nested Schema for `rules.options`

Read-only:

- **evaluation_window** (Number)
- **keep_alive** (Number)
- **max_signal_duration** (Number)


<a id="nestedobjatt--rules--query"></a>
### Nested Schema for `rules.query`

Read-only:

- **aggregation** (String)
- **distinct_fields** (List of String)
- **group_by_fields** (List of String)
- **metric** (String)
- **name** (String)
- **query** (String)


