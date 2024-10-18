---
subcategory: ""
page_title: "Metric Resource Examples"
description: |-
    Metric Resource Examples
---

### Metric Resource Examples

This page lists examples of how to create different Datadog metric types within Terraform. This list is not exhaustive and will be updated over time to provide more examples.

## Metric metadata

A metric’s metadata includes the metric name, description, and unit.

### Count metrics

```terraform
resource "datadog_metric_metadata" "request_count" {
  metric      = "request.count"
  short_name  = "Request count"
  description = "Count of requests"
  type        = "count"
  unit        = "request"
}
```

### Gauge metrics

```terraform
resource "datadog_metric_metadata" "request_time" {
  metric      = "request.time"
  short_name  = "Request time"
  description = "99th percentile request time in milliseconds"
  type        = "gauge"
  unit        = "millisecond"
}
```

### Rate metrics

```terraform
resource "datadog_metric_metadata" "request_rate" {
  metric      = "request.rate"
  short_name  = "Request rate"
  description = "Rate of requests"
  type        = "rate"
  unit        = "request"
}
```