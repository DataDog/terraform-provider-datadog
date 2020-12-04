---
page_title: "datadog_security_monitoring_rule"
---

# datadog_security_monitoring_rule Resource

Provides a Datadog [Security Monitoring Rule API](https://docs.datadoghq.com/api/v2/security-monitoring/) resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule use [datadog_security_default_rule](/resources/security_monitoring_default_rule) instead.

## Example Usage

Create a simple security monitoring rule.

```hcl
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

## Argument Reference

The following arguments are supported:

-   `message`: (Required) Message for generated signals.
-   `name`: (Required) The name of the rule.
-   `enabled`: (Optional, default = True) Whether the rule is enabled.
-   `query`: (Required) Queries for selecting logs which are part of the rule.
    -   `name`: (Optional) Name of the query.
    -   `query`: (Required) Query to run on logs.
    -   `groupByFields`: (Optional) Fields to group by.
    -   `aggregation`: (Optional) The aggregation type.Allowed values: `count`, `cardinality`, `sum`, `max`
    -   `distinctFields`: (Optional) Fields for which the cardinality is measured.
    -   `metric`: (Optional) The target field to aggregate over when using the sum or max aggregations.
-   `case`: (Required) Cases for generating signals.
    -   `name`: (Optional) Rule case name.
    -   `status`: (Required) Severity of the rule case.
    -   `condition`: (Optional) A rule case contains logical operations (`>`, `>=`, `&&`, `||`) to determine if a signal should be generated based on the event counts in the previously defined queries.
    -   `notifications`: (Optional) List of notification targets if the rule case triggers.
-   `options`: (Optional) Options on the rule
    -   `evaluation_window`: (Required) A time window is specified to match when at least one of the cases matches true.This is a sliding window and evaluates in real time. Allowed values: `0,60,300,600,900,1800,3600,7200`.
    -   `keep_alive`: (Required) Once a signal is generated, the signal will remain “open” if a case is matched at least once within this keep alive window.Allowed values: `0,60,300,600,900,1800,3600,7200,10800,21600`.
    -   `max_signal_duration`: (Required) A signal will "close" regardless of the query being matched once the time exceeds the maximum duration.This time is calculated from the first seen timestamp.Allowed values: `0,60,300,600,900,1800,3600,7200,10800,21600,43200,86400`.
-   `tags`: (Optional) Tags for generated signals.

## Import

Security monitoring rules can be imported using ID, e.g.

```console
$ terraform import datadog_security_monitoring_rule.my_monitor m0o-hto-lkb
```
