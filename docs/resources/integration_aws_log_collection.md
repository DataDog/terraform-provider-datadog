---
page_title: "datadog_integration_aws_log_collection Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog - Amazon Web Services integration log collection resource. This can be used to manage which AWS services logs are collected from for an account.
---

# Resource `datadog_integration_aws_log_collection`

Provides a Datadog - Amazon Web Services integration log collection resource. This can be used to manage which AWS services logs are collected from for an account.

## Example Usage

```terraform
# Create a new Datadog - Amazon Web Services integration lambda arn
resource "datadog_integration_aws_log_collection" "main" {
  account_id = "1234567890"
  services = ["lambda"]
}
```

## Schema

### Required

- **account_id** (String, Required) Your AWS Account ID without dashes.
- **services** (List of String, Required) A list of services to collect logs from. See the [api docs](https://docs.datadoghq.com/api/v1/aws-logs-integration/#get-list-of-aws-log-ready-services) for more details on which services are supported.

### Optional

- **id** (String, Optional) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# Amazon Web Services log collection integrations can be imported using the `account ID`.

terraform import datadog_integration_aws_log_collection.test 1234567890
```
