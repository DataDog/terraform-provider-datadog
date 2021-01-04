---
page_title: "datadog_security_monitoring_rule Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Security Monitoring Rule API resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule use datadogsecuritydefault_rule instead.
---

# Resource `datadog_security_monitoring_rule`

Provides a Datadog Security Monitoring Rule API resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule use datadog_security_default_rule instead.

## Example Usage

```terraform
resource "datadog_security_monitoring_rule" "myrule" {
    name = "My rule"

    message = "The rule has triggered."
    enabled = true

    query {
        name = "errors"
        query = "status:error"
        aggregation = "count"
        group_by_fields = ["host"]
    }

    query {
        name = "warnings"
        query = "status:warning"
        aggregation = "count"
        group_by_fields = ["host"]
    }

    case {
        status = "high"
        condition = "errors > 3 && warnings > 10"
        notifications = ["@user"]
    }

     options {
         evaluation_window = 300
         keep_alive = 600
         max_signal_duration = 900
     }

     tags = ["type:dos"]
 }
```

## Schema

### Required

- **case** (Block List, Min: 1, Max: 5) Cases for generating signals. (see [below for nested schema](#nestedblock--case))
- **message** (String) Message for generated signals.
- **name** (String) The name of the rule.
- **query** (Block List, Min: 1) Queries for selecting logs which are part of the rule. (see [below for nested schema](#nestedblock--query))

### Optional

- **enabled** (Boolean) Whether the rule is enabled.
- **id** (String) The ID of this resource.
- **options** (Block List, Max: 1) Options on rules. (see [below for nested schema](#nestedblock--options))
- **tags** (List of String) Tags for generated signals.

<a id="nestedblock--case"></a>
### Nested Schema for `case`

Required:

- **status** (String) Severity of the Security Signal.

Optional:

- **condition** (String) A rule case contains logical operations (`>`,`>=`, `&&`, `||`) to determine if a signal should be generated based on the event counts in the previously defined queries.
- **name** (String) Name of the case.
- **notifications** (List of String) Notification targets for each rule case.


<a id="nestedblock--query"></a>
### Nested Schema for `query`

Required:

- **query** (String) Query to run on logs.

Optional:

- **aggregation** (String) The aggregation type.
- **distinct_fields** (List of String) Field for which the cardinality is measured. Sent as an array.
- **group_by_fields** (List of String) Fields to group by.
- **metric** (String) The target field to aggregate over when using the sum or max aggregations.
- **name** (String) Name of the query.


<a id="nestedblock--options"></a>
### Nested Schema for `options`

Required:

- **evaluation_window** (Number) A time window is specified to match when at least one of the cases matches true. This is a sliding window and evaluates in real time.
- **keep_alive** (Number) Once a signal is generated, the signal will remain “open” if a case is matched at least once within this keep alive window.
- **max_signal_duration** (Number) A signal will “close” regardless of the query being matched once the time exceeds the maximum duration. This time is calculated from the first seen timestamp.

## Import

Import is supported using the following syntax:

```shell
# Security monitoring rules can be imported using ID, e.g.
terraform import datadog_security_monitoring_rule.my_monitor m0o-hto-lkb
```
