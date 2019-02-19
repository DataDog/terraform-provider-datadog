---
layout: "datadog"
page_title: "Datadog: datadog_metric_metadata"
sidebar_current: "docs-datadog-resource-metric_metadata"
description: |-
  Provides a Datadog metric metadata resource. This can be used to manage a metric's metadata.
---

# datadog_metric_metadata

Provides a Datadog metric_metadata resource. This can be used to manage a metric's metadata.

## Example Usage

```hcl
# Manage a Datadog metric's metadata
resource "datadog_metric_metadata" "request_time" {
  metric      = "request.time"
  short_name  = "Request time"
  description = "99th percentile request time in millseconds"
  type        = "gauge"
  unit        = "millisecond"
}
```

## Argument Reference

The following arguments are supported:

* `metric` - (Required) The name of the metric.
* `description` - (Optional) A description of the metric.
* `short_name` - (Optional) A short name of the metric.
* `unit` - (Optional) Primary unit of the metric such as 'byte' or 'operation'.
* `per_unit` - (Optional) 'Per' unit of the metric such as 'second' in 'bytes per second'.
* `statsd_interval` - (Optional) If applicable, stasd flush interval in seconds for the metric.

