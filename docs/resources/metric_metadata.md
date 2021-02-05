---
page_title: "datadog_metric_metadata Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog metric_metadata resource. This can be used to manage a metric's metadata.
---

# Resource `datadog_metric_metadata`

Provides a Datadog metric_metadata resource. This can be used to manage a metric's metadata.

## Example Usage

```terraform
# Manage a Datadog metric's metadata
resource "datadog_metric_metadata" "request_time" {
  metric      = "request.time"
  short_name  = "Request time"
  description = "99th percentile request time in millseconds"
  type        = "gauge"
  unit        = "millisecond"
}
```

## Schema

### Required

- **metric** (String, Required) The name of the metric.

### Optional

- **description** (String, Optional) A description of the metric.
- **id** (String, Optional) The ID of this resource.
- **per_unit** (String, Optional) Per unit of the metric such as `second` in `bytes per second`.
- **short_name** (String, Optional) A short name of the metric.
- **statsd_interval** (Number, Optional) If applicable, statsd flush interval in seconds for the metric.
- **type** (String, Optional) Type of the metric.
- **unit** (String, Optional) Primary unit of the metric such as `byte` or `operation`.


