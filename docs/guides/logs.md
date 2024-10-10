---
subcategory: ""
page_title: "Log Resource Examples"
description: |-
    Log Resource Examples
---

### Log Resource Examples

This page lists examples of how to create different Datadog Log Management resource types within Terraform. This list is not exhaustive and will be updated over time to provide more examples.

## Log-based metrics

### Count metrics

```terraform
resource "datadog_logs_metric" "example_count_metric" {
  name = "logs.count.metric"
  compute {
    aggregation_type = "count"
  }
  filter {
    query = "service:example"
  }
  group_by {
    path     = "@status"
    tag_name = "status"
  }
  group_by {
    path     = "@version"
    tag_name = "version"
  }
}
```

### Distribution metrics

```terraform
resource "datadog_logs_metric" "testing_logs_metric" {
  name = "testing.logs.metric"
  compute {
    aggregation_type = "distribution"
    path             = "@duration"
  }
  filter {
    query = "service:test"
  }
  group_by {
    path     = "@status"
    tag_name = "status"
  }
  group_by {
    path     = "@version"
    tag_name = "version"
  }
}
```