---
page_title: "datadog_logs_integration_pipeline Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Pipeline API resource to manage the integrations.
  Integration pipelines are the pipelines that are automatically installed for your organization when sending the logs with specific sources. You don't need to maintain or update these types of pipelines. Keeping them as resources, however, allows you to manage the order of your pipelines by referencing them in your datadoglogspipelineorder resource. If you don't need the pipelineorder feature, this resource declaration can be omitted.
---

# Resource `datadog_logs_integration_pipeline`

Provides a Datadog Logs Pipeline API resource to manage the integrations.

Integration pipelines are the pipelines that are automatically installed for your organization when sending the logs with specific sources. You don't need to maintain or update these types of pipelines. Keeping them as resources, however, allows you to manage the order of your pipelines by referencing them in your datadog_logs_pipeline_order resource. If you don't need the pipeline_order feature, this resource declaration can be omitted.

## Example Usage

```terraform
resource "datadog_logs_integration_pipeline" "python" {
    is_enabled = true
}
```

## Schema

### Optional

- **id** (String, Optional) The ID of this resource.
- **is_enabled** (Boolean, Optional)

## Import

Import is supported using the following syntax:

```shell
terraform import <resource.name> <pipelineID>
```
