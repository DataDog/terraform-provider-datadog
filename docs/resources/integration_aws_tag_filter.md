---
page_title: "datadog_integration_aws_tag_filter Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog AWS tag filter resource. This can be used to create and manage Datadog AWS tag filters - US site’s endpoint only
---

# Resource `datadog_integration_aws_tag_filter`

Provides a Datadog AWS tag filter resource. This can be used to create and manage Datadog AWS tag filters - US site’s endpoint only

## Example Usage

```terraform
resource "datadog_integration_aws_tag_filter" "foo" {
  account_id = "123456789010"
  namespace = "sqs"
  tag_filter_str = "key:value"
}
```

## Schema

### Required

- **account_id** (String) Your AWS Account ID without dashes.
- **namespace** (String) The namespace associated with the tag filter entry. Allowed enum values: 'elb', 'application_elb', 'sqs', 'rds', 'custom', 'network_elb,lambda'
- **tag_filter_str** (String) The tag filter string.

### Optional

- **id** (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# Amazon Web Services log filter resource can be imported using their account ID and namespace separated with a colon (:).
terraform import datadog_integration_aws_tag_filter.foo ${account_id}:${namespace}
```
