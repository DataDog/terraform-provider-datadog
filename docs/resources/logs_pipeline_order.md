---
page_title: "datadog_logs_pipeline_order Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Pipeline API resource, which is used to manage Datadog log pipelines order.
---

# Resource `datadog_logs_pipeline_order`

Provides a Datadog Logs Pipeline API resource, which is used to manage Datadog log pipelines order.

## Example Usage

```terraform
resource "datadog_logs_pipeline_order" "sample_pipeline_order" {
    name = "sample_pipeline_order"
    depends_on = [
        "datadog_logs_custom_pipeline.sample_pipeline",
        "datadog_logs_integration_pipeline.python"
    ]
    pipelines = [
        "${datadog_logs_custom_pipeline.sample_pipeline.id}",
        "${datadog_logs_integration_pipeline.python.id}"
    ]
}
```

## Schema

### Required

- **name** (String, Required)
- **pipelines** (List of String, Required)

### Optional

- **id** (String, Optional) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# There must be at most one datadog_logs_pipeline_order resource. Pipeline order creation is not supported from logs config API. You can import the datadog_logs_pipeline_order or create a pipeline order (which is actually doing the update operation).

terraform import <datadog_logs_pipeline_order.name> <name>
```
