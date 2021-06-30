---
page_title: "datadog_security_monitoring_rule Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Security Monitoring Rule API resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule use datadog_security_default_rule instead.
---

# Resource `datadog_security_monitoring_rule`

Provides a Datadog Security Monitoring Rule API resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule use `datadog_security_default_rule` instead.

## Example Usage

```terraform
resource "datadog_security_monitoring_rule" "myrule" {
  name = "My rule"

  message = "The rule has triggered."
  enabled = true
  has_extended_title = true


  query {
    name            = "errors"
    query           = "status:error"
    aggregation     = "count"
    group_by_fields = ["host"]
  }

  query {
    name            = "warnings"
    query           = "status:warning"
    aggregation     = "count"
    group_by_fields = ["host"]
  }

  case {
    status        = "high"
    condition     = "errors > 3 && warnings > 10"
    notifications = ["@user"]
  }

  options {
    detection_method    = "threshold"
    evaluation_window   = 300
    keep_alive          = 600
    max_signal_duration = 900
  }

  tags = ["type:dos"]
}
```

## Schema

### Required

- **case** (Block List, Min: 1, Max: 5) Cases for generating signals. (see [below for nested schema](#nestedblock--case))
- **message** (String, Required) Message for generated signals.
- **name** (String, Required) The name of the rule.
- **query** (Block List, Min: 1) Queries for selecting logs which are part of the rule. (see [below for nested schema](#nestedblock--query))

### Optional

- **enabled** (Boolean, Optional) Whether the rule is enabled.
- **has_extended_title** (Boolean, Optional) Whether the notifications include the triggering group-by values in their title.
- **id** (String, Optional) The ID of this resource.
- **options** (Block List, Max: 1) Options on rules. (see [below for nested schema](#nestedblock--options))
- **filter** (Block List, Optional) Additional queries to filter matched events before they are processed. (see [below for nested schema](#nestedblock--filter))
- **tags** (List of String, Optional) Tags for generated signals.


<a id="nestedblock--case"></a>
### Nested Schema for `case`

Required:

- **status** (String, Required) Severity of the Security Signal.

Optional:

- **condition** (String, Optional) A rule case contains logical operations (`>`,`>=`, `&&`, `||`) to determine if a signal should be generated based on the event counts in the previously defined queries.
- **name** (String, Optional) Name of the case.
- **notifications** (List of String, Optional) Notification targets for each rule case.


<a id="nestedblock--query"></a>
### Nested Schema for `query`

Required:

- **query** (String, Required) Query to run on logs.

Optional:

- **aggregation** (String, Optional) The aggregation type.
- **distinct_fields** (List of String, Optional) Field for which the cardinality is measured. Sent as an array.
- **group_by_fields** (List of String, Optional) Fields to group by.
- **metric** (String, Optional) The target field to aggregate over when using the sum or max aggregations.
- **name** (String, Optional) Name of the query.

<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

Required:

- **query** (String, Required) Query to run on logs.
- **action** (String, Required) The type of filtering action (require or suppress).


<a id="nestedblock--options"></a>
### Nested Schema for `options`

Required:

- **evaluation_window** (Number, Required) A time window is specified to match when at least one of the cases matches true. This is a sliding window and evaluates in real time.
- **keep_alive** (Number, Required) Once a signal is generated, the signal will remain “open” if a case is matched at least once within this keep alive window.
- **max_signal_duration** (Number, Required) A signal will “close” regardless of the query being matched once the time exceeds the maximum duration. This time is calculated from the first seen timestamp.

Optional:

- **detection_method** (String, Optional) The detection method. Default to `threshold`.
- **new_value_options** (Block List, Max: 1) Specific options for `new_value` detection method. (see [below for nested schema](#nestedblock--new-value-options))

<a id="nestedblock--options"></a>
### Nested Schema for `new_value_options`

Required:

- **forget_after** (Number, Required) The duration in days after which a learned value is forgotten.
- **learning_duration** (Number, Required) The duration in days during which values are learned, and after which signals will be generated for values that weren't learned. If set to 0, a signal will be generated for all new values after the first value is learned.

## Import

Import is supported using the following syntax:

```shell
# Security monitoring rules can be imported using ID, e.g.
terraform import datadog_security_monitoring_rule.my_rule m0o-hto-lkb
```
