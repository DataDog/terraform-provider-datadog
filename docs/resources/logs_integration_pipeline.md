---
page_title: "datadog_logs_integration_pipeline"
---

# datadog_logs_integration_pipeline Resource

Provides a Datadog [Logs Pipeline API](https://docs.datadoghq.com/api/v1/logs-pipelines/) resource to manage
the [integrations](https://docs.datadoghq.com/logs/log_collection/?tab=tcpussite).

Integration pipelines are the pipelines that are automatically installed for your organization when sending the logs with
specific sources. You don't need to maintain or update these types of pipelines. Keeping them as resources, however,
allows you to manage the order of your pipelines by referencing them in your
[datadog_logs_pipeline_order](logs_pipeline_order.html#datadog_logs_pipeline_order) resource. If you don't need the
`pipeline_order` feature, this resource declaration can be omitted.

## Example Usage

```hcl
resource "datadog_logs_integration_pipeline" "python" {
    is_enabled = true
}
```

## Argument Reference

* `is_enabled`: (Required) Boolean value to enable your pipeline.

`is_enabled` is the only value that can be modified for integration pipeline.

## Import

```terraform import <resource.name> <pipelineID>```
