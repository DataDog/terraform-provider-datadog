---
page_title: "security_monitoring_rules"
---

# datadog_security_monitoring_rules Data Source

Use this data source to retrieve information about existing security monitoring rules for use in other resources.

## Example Usage

```
data "datadog_security_monitoring_rules" "test" {
  name_filter = "attack"
  tags_filter = ["foo:bar"]
  default_only_filter = true
}
```

## Argument Reference

-   `name_filter`: (Optional) A rule name to limit the search.
-   `tags_filter`: (Optional) A list of rule tags to limit the search.
-   `default_only_filter`: (Optional) Limit search to default rules.
-   `user_only_filter`: (Optional) Limit search to user rules.

## Attributes Reference

-   `rule_ids`: List of ids of the matched rules.
-   `rules`: List of rules.
    -   `case`: Cases for generating signals.
        -   `name`: Rule case name.
        -   `status`: Severity of the rule case.
        -   `condition`: A rule case contains logical operations (`>`,`>=`, `&&`, `||`) to determine if a signal should be generated based on the event counts in the previously defined queries.
        -   `notifications`: List of notification targets if the rule case triggers.
    -   `enabled`: Whether the rule is enabled.
    -   `message`: Message for generated signals.
    -   `name`: The name of the rule.
    -   `options`: Options on the rule
        -   `evaluation_window`: A time window is specified to match when at least one of the cases matches true. This is a sliding window and evaluates in real time. Allowed values: `0,60,300,600,900,1800,3600,7200`.
        -   `keep_alive`: Once a signal is generated, the signal will remain “open” if a case is matched at least once within this keep alive window. Allowed values: `0,60,300,600,900,1800,3600,7200,10800,21600`.
        -   `max_signal_duration`: A signal will "close" regardless of the query being matched once the time exceeds the maximum duration. This time is calculated from the first seen timestamp. Allowed values: `0,60,300,600,900,1800,3600,7200,10800,21600,43200,86400`.
    -   `query`: Queries for selecting logs which are part of the rule.
        -   `aggregation`: The aggregation type. Allowed values: `count`, `cardinality`, `sum`, `max`
        -   `distinctFields`: Fields for which the cardinality is measured.
        -   `groupByFields`: Fields to group by.
        -   `metric`: The target field to aggregate over when using the sum or max aggregations.
        -   `name`: Name of the query.
        -   `query`: Query to run on logs.
    -   `tags`: Tags for generated signals.
