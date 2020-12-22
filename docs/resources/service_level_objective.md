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


# Monitor-Based SLO
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

## Schema

### Required

- **name** (String, Required)
- **thresholds** (Block List, Min: 1) (see [below for nested schema](#nestedblock--thresholds))
- **type** (String, Required)

### Optional

- **description** (String, Optional)
- **force_delete** (Boolean, Optional)
- **groups** (Set of String, Optional) A static set of groups to filter monitor-based SLOs
- **id** (String, Optional) The ID of this resource.
- **monitor_ids** (Set of Number, Optional) A static set of monitor IDs to use as part of the SLO
- **monitor_search** (String, Optional)
- **query** (Block List, Max: 1) The metric query of good / total events (see [below for nested schema](#nestedblock--query))
- **tags** (Set of String, Optional)
- **validate** (Boolean, Optional)

<a id="nestedblock--thresholds"></a>
### Nested Schema for `thresholds`

Required:

- **target** (Number, Required)
- **timeframe** (String, Required)

Optional:

- **target_display** (String, Optional)
- **warning** (Number, Optional)
- **warning_display** (String, Optional)


<a id="nestedblock--query"></a>
### Nested Schema for `query`

Required:

- **denominator** (String, Required)
- **numerator** (String, Required)

## Import

Import is supported using the following syntax:

```shell
# Service Level Objectives can be imported using their string ID, e.g.
terraform import datadog_service_level_objective.baz 12345678901234567890123456789012
```
