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

- **metric** (String, Required)

### Optional

- **description** (String, Optional)
- **id** (String, Optional) The ID of this resource.
- **per_unit** (String, Optional)
- **short_name** (String, Optional)
- **statsd_interval** (Number, Optional)
- **type** (String, Optional)
- **unit** (String, Optional)


