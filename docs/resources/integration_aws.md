---
page_title: "datadog_integration_aws Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog - Amazon Web Services integration resource. This can be used to create and manage Datadog - Amazon Web Services integration.
---

# Resource `datadog_integration_aws`

Provides a Datadog - Amazon Web Services integration resource. This can be used to create and manage Datadog - Amazon Web Services integration.

## Example Usage

```terraform
# Create a new Datadog - Amazon Web Services integration
resource "datadog_integration_aws" "sandbox" {
    account_id = "1234567890"
    role_name = "DatadogAWSIntegrationRole"
    filter_tags = ["key:value"]
    host_tags = ["key:value", "key2:value2"]
    account_specific_namespace_rules = {
        auto_scaling = false
        opsworks = false
    }
    excluded_regions = ["us-east-1", "us-west-2"]
}
```

## Schema

### Required

- **account_id** (String, Required) Your AWS Account ID without dashes.
- **role_name** (String, Required) Your Datadog role delegation name.

### Optional

- **account_specific_namespace_rules** (Map of Boolean, Optional) Enables or disables metric collection for specific AWS namespaces for this AWS account only. A list of namespaces can be found at the [available namespace rules API endpoint](https://docs.datadoghq.com/api/v1/aws-integration/#list-namespace-rules).
- **excluded_regions** (List of String, Optional) An array of AWS regions to exclude from metrics collection.
- **filter_tags** (List of String, Optional) Array of EC2 tags (in the form `key:value`) defines a filter that Datadog uses when collecting metrics from EC2. Wildcards, such as `?` (for single characters) and `*` (for multiple characters) can also be used. Only hosts that match one of the defined tags will be imported into Datadog. The rest will be ignored. Host matching a given tag can also be excluded by adding `!` before the tag. e.x. `env:production,instance-type:c1.*,!region:us-east-1`.
- **host_tags** (List of String, Optional) Array of tags (in the form `key:value`) to add to all hosts and metrics reporting through this integration.
- **id** (String, Optional) The ID of this resource.

### Read-only

- **external_id** (String, Read-only) AWS External ID. **NOTE** This provider will not be able to detect changes made to the `external_id` field from outside Terraform.

## Import

Import is supported using the following syntax:

```shell
# Amazon Web Services integrations can be imported using their account ID and role name separated with a colon (:), while the external_id should be passed by setting an environment variable called EXTERNAL_ID

EXTERNAL_ID=${external_id} terraform import datadog_integration_aws.test ${account_id}:${role_name}
```
