---
layout: "datadog"
page_title: "Datadog: datadog_logs_pipeline_order"
sidebar_current: "docs-datadog-resource-logs-pipeline-order"
description: |-
  Provides a Datadog logs pipeline order resource. This can be used to manage the order of logs pipelines.
---

# datadog_logs_pipeline_order

Provides a Datadog [Logs Pipeline API](https://docs.datadoghq.com/api/?lang=python#logs-pipelines) resource. This can be used to manage Datadog logs pipelines order.


## Example Usage

Create a Datadog logs pipeline order resource.

```hcl
resource "datadog_logs_pipeline_order" "sample_pipeline_order" {
    name = "sample_pipeline_order"
    depends_on = [
        "datadog_logs_pipeline.sample_pipeline"
    ]
    pipelines = [
        "${datadog_logs_pipeline.sample_pipeline.id}"
    ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name attribute in the resource `datadog_logs_pipeline_order` needs to be unique. It's better to set to the same value as the resource `NAME`. 
There is no related field can be found in  [Logs Pipeline API](https://docs.datadoghq.com/api/?lang=python#get-pipeline-order).
* `pipelines` - (Required) The pipeline ids list. The order of pipeline ids in this attribute defines the overall pipelines order for Logs.

## Attributes Reference

* `pipelines` - The `pipelines` list contains the ids of resources [datadog_logs_pipeline](logs_pipeline.html#datadog_logs_pipeline) that are created and imported.
Update the order of pipelines in this list reflects the application order of the pipelines. You cannot delete or create pipelines by deleting or adding ids to this list.

## Import

There must be at most one `datadog_logs_pipeline_order` resource. Pipeline order creation is not supported from logs config API. 
You should import the `datadog_logs_pipeline_order` rather than create a pipeline order.

```
terraform import <datadog_logs_pipeline_order.name> <name>
```
