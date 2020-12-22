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

- **account_id** (String, Required)
- **lambda_arn** (String, Required)

### Optional

- **id** (String, Optional) The ID of this resource.


