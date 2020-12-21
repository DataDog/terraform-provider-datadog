---
page_title: "datadog_integration_aws_log_collection Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog - Amazon Web Services integration log collection resource. This can be used to manage which AWS services logs are collected from for an account.
---

# Resource `datadog_integration_aws_log_collection`

Provides a Datadog - Amazon Web Services integration log collection resource. This can be used to manage which AWS services logs are collected from for an account.



## Schema

### Required

- **account_id** (String, Required)
- **services** (List of String, Required)

### Optional

- **id** (String, Optional) The ID of this resource.


