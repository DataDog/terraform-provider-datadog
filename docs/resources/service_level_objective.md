---
page_title: "datadog_service_level_objective"
---

# datadog_service_level_objective Resource

Provides a Datadog service level objective resource. This can be used to create and manage Datadog service level objectives.

## Example Usage

### Metric-Based SLO
```hcl
# Create a new Datadog service level objective
resource "datadog_service_level_objective" "foo" {
  name               = "Example Metric SLO"
  type               = "metric"
  description        = "My custom metric SLO"
  query {
    numerator = "sum:my.custom.count.metric{type:good_events}.as_count()"
    denominator = "sum:my.custom.count.metric{*}.as_count()"
  }

  thresholds {
    timeframe = "7d"
    target = 99.9
    warning = 99.99
    target_display = "99.900"
    warning_display = "99.990"
  }

  thresholds {
    timeframe = "30d"
    target = 99.9
    warning = 99.99
    target_display = "99.900"
    warning_display = "99.990"
  }

  tags = ["foo:bar", "baz"]
}
```

### Monitor-Based SLO
```hcl
# Create a new Datadog service level objective
resource "datadog_service_level_objective" "bar" {
  name               = "Example Monitor SLO"
  type               = "monitor"
  description        = "My custom monitor SLO"
  monitor_ids = [1, 2, 3]

  thresholds {
    timeframe = "7d"
    target = 99.9
    warning = 99.99
  }

  thresholds {
    timeframe = "30d"
    target = 99.9
    warning = 99.99
  }

  tags = ["foo:bar", "baz"]
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Required) The type of the service level objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation](https://docs.datadoghq.com/api/v1/service-level-objectives/#create-a-slo-object) page. Available options to choose from are:
    * `metric`
    * `monitor`
* `name` - (Required) Name of Datadog service level objective
* `description` - (Optional) A description of this service level objective.
* `tags` (Optional) A list of tags to associate with your service level objective. This can help you categorize and filter service level objectives in the service level objectives page of the UI. Note: it's not currently possible to filter by these tags when querying via the API
* `thresholds` - (Required) - A list of thresholds and targets that define the service level objectives from the provided SLIs.
    * `timeframe` (Required) - the time frame for the objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation](https://docs.datadoghq.com/api/v1/service-level-objectives/#create-a-slo-object) page. Available options to choose from are:
        * `7d`
        * `30d`
        * `90d`
    * `target` - (Required) the objective's target `[0,100]`
    * `target_display` - (Optional) the string version to specify additional digits in the case of `99` but want 3 digits like `99.000` to display.
    * `warning` - (Optional) the objective's warning value `[0,100]`. This must be `> target` value.
    * `warning_display` - (Optional) the string version to specify additional digits in the case of `99` but want 3 digits like `99.000` to display.

The following options are specific to the `type` of service level objective:

* `metric` type SLOs:
    * `query` - (Required) The metric query configuration to use for the SLI. This is a dictionary and requires both the `numerator` and `denominator` fields which should be `count` metrics using the `sum` aggregator.
        * `numerator` - (Required) the sum of all the `good` events
        * `denominator` - (Required) the sum of the `total` events
        * Example Usage:

```hcl
query {
    numerator   = "sum:my.custom.count.metric{type:good}.as_count()"
    denominator = "sum:my.custom.count.metric{*}.as_count()"
}
```

* `monitor` type SLOs:
    * `monitor_ids` - (Required) A list of numeric monitor IDs for which to use as SLIs. Their tags will be auto-imported into `monitor_tags` field in the API resource.
    * `groups` - (Optional) A custom set of groups from the monitor(s) for which to use as the SLI instead of all the groups.


## Attributes Reference

The following attributes are exported:

* `id` - ID of the Datadog service level objective

## Import

Service Level Objectives can be imported using their string ID, e.g.

```
$ terraform import datadog_service_level_objective.baz 12345678901234567890123456789012
```
