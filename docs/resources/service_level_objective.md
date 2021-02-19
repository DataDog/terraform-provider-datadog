---
page_title: "datadog_service_level_objective Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog service level objective resource. This can be used to create and manage Datadog service level objectives.
---

# Resource `datadog_service_level_objective`

Provides a Datadog service level objective resource. This can be used to create and manage Datadog service level objectives.

## Example Usage

```terraform
# Metric-Based SLO
# Create a new Datadog service level objective
resource "datadog_service_level_objective" "foo" {
  name        = "Example Metric SLO"
  type        = "metric"
  description = "My custom metric SLO"
  query {
    numerator   = "sum:my.custom.count.metric{type:good_events}.as_count()"
    denominator = "sum:my.custom.count.metric{*}.as_count()"
  }

  thresholds {
    timeframe       = "7d"
    target          = 99.9
    warning         = 99.99
    target_display  = "99.900"
    warning_display = "99.990"
  }

  thresholds {
    timeframe       = "30d"
    target          = 99.9
    warning         = 99.99
    target_display  = "99.900"
    warning_display = "99.990"
  }

  tags = ["foo:bar", "baz"]
}


# Monitor-Based SLO
# Create a new Datadog service level objective
resource "datadog_service_level_objective" "bar" {
  name        = "Example Monitor SLO"
  type        = "monitor"
  description = "My custom monitor SLO"
  monitor_ids = [1, 2, 3]

  thresholds {
    timeframe = "7d"
    target    = 99.9
    warning   = 99.99
  }

  thresholds {
    timeframe = "30d"
    target    = 99.9
    warning   = 99.99
  }

  tags = ["foo:bar", "baz"]
}
```

## Schema

### Required

- **name** (String, Required) Name of Datadog service level objective
- **thresholds** (Block List, Min: 1) A list of thresholds and targets that define the service level objectives from the provided SLIs. (see [below for nested schema](#nestedblock--thresholds))
- **type** (String, Required) The type of the service level objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/service-level-objectives/#create-a-slo-object). Available options to choose from are: `metric` and `monitor`.

### Optional

- **description** (String, Optional) A description of this service level objective.
- **force_delete** (Boolean, Optional) A boolean indicating whether this monitor can be deleted even if it’s referenced by other resources (e.g. dashboards).
- **groups** (Set of String, Optional) A static set of groups to filter monitor-based SLOs
- **id** (String, Optional) The ID of this resource.
- **monitor_ids** (Set of Number, Optional) A static set of monitor IDs to use as part of the SLO
- **monitor_search** (String, Optional)
- **query** (Block List, Max: 1) The metric query of good / total events (see [below for nested schema](#nestedblock--query))
- **tags** (Set of String, Optional) A list of tags to associate with your service level objective. This can help you categorize and filter service level objectives in the service level objectives page of the UI. Note: it's not currently possible to filter by these tags when querying via the API
- **validate** (Boolean, Optional) Whether or not to validate the SLO.

<a id="nestedblock--thresholds"></a>
### Nested Schema for `thresholds`

Required:

- **target** (Number, Required) The objective's target in`[0,100]`.
- **timeframe** (String, Required) The time frame for the objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API documentation page. Available options to choose from are: `7d`, `30d`, `90d`.

Optional:

- **target_display** (String, Optional) A string representation of the target that indicates its precision. It uses trailing zeros to show significant decimal places (e.g. `98.00`).
- **warning** (Number, Optional) The objective's warning value in `[0,100]`. This must be greater than the target value.
- **warning_display** (String, Optional) A string representation of the warning target (see the description of the target_display field for details).


<a id="nestedblock--query"></a>
### Nested Schema for `query`

Required:

- **denominator** (String, Required) The sum of the `total` events.
- **numerator** (String, Required) The sum of all the `good` events.

## Import

Import is supported using the following syntax:

```shell
# Service Level Objectives can be imported using their string ID, e.g.
terraform import datadog_service_level_objective.baz 12345678901234567890123456789012
```
