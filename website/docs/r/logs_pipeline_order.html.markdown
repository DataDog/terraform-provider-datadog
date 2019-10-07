---
layout: "datadog"
page_title: "Datadog: datadog_logs_pipeline_order"
sidebar_current: "docs-datadog-resource-logs-pipeline-order"
description: |-
  Provides a Datadog logs pipeline order resource, which is used to manage the order of log pipelines.
---

# datadog_logs_pipeline_order

Provides a Datadog [Logs Pipeline API](https://docs.datadoghq.com/api/?lang=python#logs-pipelines) resource, which is used to manage Datadog log pipelines order.


## Example Usage

```hcl
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

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name attribute in the resource `datadog_logs_pipeline_order` needs to be unique. It's recommended to use the same value as the resource `NAME`. 
No related field is available in  [Logs Pipeline API](https://docs.datadoghq.com/api/?lang=python#get-pipeline-order).
* `pipelines` - (Required) The pipeline IDs list. The order of pipeline IDs in this attribute defines the overall pipeline order for logs.

## Attributes Reference

* `pipelines` - The `pipelines` list contains the IDs of resources created and imported by the 
[datadog_logs_custom_pipeline](logs_custom_pipeline.html#datadog_logs_custom_pipeline) and 
[datadog_logs_integration_pipeline](logs_integration_pipeline.html#datadog_logs_integration_pipeline).
Updating the order of pipelines in this list reflects the application order of the pipelines. You cannot delete or create pipeline by deleting or adding IDs to this list.

## Import

There must be at most one `datadog_logs_pipeline_order` resource. Pipeline order creation is not supported from logs config API. 
You can import the `datadog_logs_pipeline_order` or create a pipeline order (which is actually doing the update operation).

```
terraform import <datadog_logs_pipeline_order.name> <name>
```

## Troubleshooting

Sometimes when applying your `datadog_logs_pipeline_order` resource, you can get an error with status `422` 
`Unprocessable Entity` with error message like `Cannot map pipelines to existing ones`, most likely the pipelines are 
incompatible with the ones declared in your `datadog_logs_pipeline_order` resource. In this case, make
sure that no pipeline gets added or deleted via other method (for example, through API call or from Datadog Logs
configuration UI).
