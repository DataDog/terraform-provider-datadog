---
layout: "datadog"
page_title: "Datadog: datadog_integration_aws"
sidebar_current: "docs-datadog-resource-integration_aws"
description: |-
  Provides a Datadog - Amazon Web Services integration resource. This can be used to create and manage the integrations.
---

# datadog_integration_aws

Provides a Datadog - Amazon Web Services integration resource. This can be used to create and manage Datadog - Amazon Web Services integration.

Update operations are currently not supported with datadog API so any change forces a new resource.

## Example Usage

```hcl
# Create a new Datadog - Amazon Web Services integration
resource "datadog_integration_aws" "sandbox" {
    account_id = "1234567890"
    role_name = "DatadogAWSIntegrationRole"
    filter_tags = ["key:value"]
    host_tags = ["key:value", "key2:value2"]
    account_specific_namespace_rules = {
        "auto_scaling" = false
        "opsworks" = false
    }
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) Your AWS Account ID without dashes.
* `role_name` - (Required) Your Datadog role delegation name.
* `filter_tags` - (Optional) Array of EC2 tags (in the form `key:value`) defines a filter that Datadog use when collecting metrics from EC2. Wildcards, such as `?` (for single characters) and `*` (for multiple characters) can also be used.
  
  Only hosts that match one of the defined tags will be imported into Datadog. The rest will be ignored. Host matching a given tag can also be excluded by adding `!` before the tag.
  
  e.x. `env:production,instance-type:c1.*,!region:us-east-1`

* `host_tags` - (Optinal) Array of tags (in the form key:value) to add to all hosts and metrics reporting through this integration.
* `account_specific_namespace_rules` - (Optinal) Enables or disables metric collection for specific AWS namespaces for this AWS account only. A list of namespaces can be found at the [available namespace rules API endpoint](https://api.datadoghq.com/api/v1/integration/aws/available_namespace_rules).

### See also
* [Datadog API Reference > Integrations > AWS](https://docs.datadoghq.com/api/?lang=bash#aws)

## Attributes Reference

The following attributes are exported:

* `external_id` - AWS External ID

## Import

Amazon Web Services integrations can be imported using their account ID and role name, e.g.

```
$ EXTERNAL_ID=${external_id} terraform import datadog_integration_aws.test ${account_id}_${role_name}
```

