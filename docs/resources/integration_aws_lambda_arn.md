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



## Schema

### Required

- **account_id** (String, Required)
- **lambda_arn** (String, Required)

### Optional

- **id** (String, Optional) The ID of this resource.


