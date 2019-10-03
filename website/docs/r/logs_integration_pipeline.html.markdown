---
layout: "datadog"
page_title: "Datadog: datadog_logs_integration_pipeline"
sidebar_current: "docs-datadog-resource-logs-integration-pipeline"
description: |-
  Provides a Datadog logs integration pipeline resource, which is used to create and manage logs integration pipelines.
---

# datadog_logs_integration_pipeline

Provides a Datadog [Logs Pipeline API](https://docs.datadoghq.com/api/?lang=python#logs-pipelines) resource to manage
the [integrations](https://docs.datadoghq.com/logs/log_collection/?tab=tcpussite).

Integration pipelines are the pipelines that automatically installed for your organization when sending the logs with 
specific sources. You don't need to maintain or update these types of pipelines. Keeping them as resources, however, 
allows you to manage the order of your pipelines by referencing them in your 
[datadog_logs_pipeline_order](logs_pipeline_order.html#datadog_logs_pipeline_order). If you don't need `pipeline_order`
feature, this resource declaration can be omitted.

## Example Usage

```hcl
resource "datadog_logs_integration_pipeline" "python" {
    is_enabled = true
}
```

## Argument Reference

* `is_enabled` - (Required) Boolean value to enable your pipeline.

`is_enabled` is the only value that can be modified for integration pipeline.

## Import

```terraform import <resource.name> <pipelineID>``` 
