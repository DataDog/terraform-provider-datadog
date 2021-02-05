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

- **default_only_filter** (Boolean, Optional) Limit the search to default rules
- **id** (String, Optional) The ID of this resource.
- **name_filter** (String, Optional) A rule name to limit the search
- **tags_filter** (List of String, Optional) A list of tags to limit the search
- **user_only_filter** (Boolean, Optional) Limit the search to user rules

### Read-only

- **rule_ids** (List of String, Read-only) List of IDs of the matched rules.
- **rules** (Block List) List of rules. (see [below for nested schema](#nestedblock--rules))

<a id="nestedblock--rules"></a>
### Nested Schema for `rules`

Required:

- **case** (Block List, Min: 1, Max: 5) Cases for generating signals. (see [below for nested schema](#nestedblock--rules--case))
- **message** (String, Required) Message for generated signals.
- **name** (String, Required) The name of the rule.
- **query** (Block List, Min: 1) Queries for selecting logs which are part of the rule. (see [below for nested schema](#nestedblock--rules--query))

Optional:

- **enabled** (Boolean, Optional) Whether the rule is enabled.
- **options** (Block List, Max: 1) Options on rules. (see [below for nested schema](#nestedblock--rules--options))
- **tags** (List of String, Optional) Tags for generated signals.

<a id="nestedblock--rules--case"></a>
### Nested Schema for `rules.case`

Required:

- **status** (String, Required) Severity of the Security Signal.

Optional:

- **condition** (String, Optional) A rule case contains logical operations (`>`,`>=`, `&&`, `||`) to determine if a signal should be generated based on the event counts in the previously defined queries.
- **name** (String, Optional) Name of the case.
- **notifications** (List of String, Optional) Notification targets for each rule case.


<a id="nestedblock--rules--query"></a>
### Nested Schema for `rules.query`

Required:

- **query** (String, Required) Query to run on logs.

Optional:

- **aggregation** (String, Optional) The aggregation type.
- **distinct_fields** (List of String, Optional) Field for which the cardinality is measured. Sent as an array.
- **group_by_fields** (List of String, Optional) Fields to group by.
- **metric** (String, Optional) The target field to aggregate over when using the sum or max aggregations.
- **name** (String, Optional) Name of the query.


<a id="nestedblock--rules--options"></a>
### Nested Schema for `rules.options`

Required:

- **evaluation_window** (Number, Required) A time window is specified to match when at least one of the cases matches true. This is a sliding window and evaluates in real time.
- **keep_alive** (Number, Required) Once a signal is generated, the signal will remain “open” if a case is matched at least once within this keep alive window.
- **max_signal_duration** (Number, Required) A signal will “close” regardless of the query being matched once the time exceeds the maximum duration. This time is calculated from the first seen timestamp.


