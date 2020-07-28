---
layout: "datadog"
page_title: "Datadog: datadog_integration_aws_lambda_arn"
sidebar_current: "docs-datadog-resource-integration_aws_lambda_arn"
description: |-
  Provides a Datadog - Amazon Web Services integration Lambda ARN resource. This can be used to create and manage the
  log collection Lambdas for an account.
---

# datadog_integration_aws_lambda_arn

Provides a Datadog - Amazon Web Services integration Lambda ARN resource. This can be used to create and manage the
log collection Lambdas for an account.

Update operations are currently not supported with datadog API so any change forces a new resource.

## Example Usage

```hcl
# Create a new Datadog - Amazon Web Services integration Lambda ARN

resource "datadog_integration_aws_lambda_arn" "main_collector" {
  account_id = "1234567890"
  lambda_arn = "arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) Your AWS Account ID without dashes.
* `lambda_arn` - (Required) The ARN of the Datadog forwarder Lambda.

### See also
* [Datadog API Reference > Integrations > AWS](https://docs.datadoghq.com/api/v1/aws-integration/)
* [Datadog log forwarder](https://github.com/DataDog/datadog-serverless-functions/tree/master/aws/logs_monitoring)

## Attributes Reference

All provided arguments are exported.

## Import

Amazon Web Services Lambda ARN integrations can be imported using their `account_id` and `lambda_arn` separated with a space (` `).

```
$ terraform import datadog_integration_aws_lambda_arn.test "1234567890 arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"
```
