---
page_title: "datadog_integration_aws_lambda_arn Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog - Amazon Web Services integration Lambda ARN resource. This can be used to create and manage the log collection Lambdas for an account.
  Update operations are currently not supported with datadog API so any change forces a new resource.
---

# Resource `datadog_integration_aws_lambda_arn`

Provides a Datadog - Amazon Web Services integration Lambda ARN resource. This can be used to create and manage the log collection Lambdas for an account.

Update operations are currently not supported with datadog API so any change forces a new resource.

## Example Usage

```terraform
# Create a new Datadog - Amazon Web Services integration Lambda ARN
resource "datadog_integration_aws_lambda_arn" "main_collector" {
  account_id = "1234567890"
  lambda_arn = "arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"
}
```

## Schema

### Required

- **account_id** (String, Required) Your AWS Account ID without dashes.
- **lambda_arn** (String, Required) The ARN of the Datadog forwarder Lambda.

### Optional

- **id** (String, Optional) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# Amazon Web Services Lambda ARN integrations can be imported using their account_id and lambda_arn separated with a space (` `).
terraform import datadog_integration_aws_lambda_arn.test "1234567890 arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"
```
