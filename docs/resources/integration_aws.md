---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_integration_aws Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  !>This resource is deprecated - use the datadog_integration_aws_account resource instead.
  Provides a Datadog - Amazon Web Services integration resource. This can be used to create and manage Datadog - Amazon Web Services integration.
---

# datadog_integration_aws (Resource)

!>This resource is deprecated - use the `datadog_integration_aws_account` resource instead.

Provides a Datadog - Amazon Web Services integration resource. This can be used to create and manage Datadog - Amazon Web Services integration.

## Example Usage

```terraform
# Create a new Datadog - Amazon Web Services integration
resource "datadog_integration_aws" "sandbox" {
  account_id  = "1234567890"
  role_name   = "DatadogAWSIntegrationRole"
  filter_tags = ["key:value"]
  host_tags   = ["key:value", "key2:value2"]
  account_specific_namespace_rules = {
    auto_scaling = false
    opsworks     = false
  }
  excluded_regions = ["us-east-1", "us-west-2"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `access_key_id` (String) Your AWS access key ID. Only required if your AWS account is a GovCloud or China account.
- `account_id` (String) Your AWS Account ID without dashes.
- `account_specific_namespace_rules` (Map of Boolean) Enables or disables metric collection for specific AWS namespaces for this AWS account only. A list of namespaces can be found at the [available namespace rules API endpoint](https://docs.datadoghq.com/api/v1/aws-integration/#list-namespace-rules).
- `cspm_resource_collection_enabled` (String) Whether Datadog collects cloud security posture management resources from your AWS account. This includes additional resources not covered under the general resource_collection.
- `excluded_regions` (Set of String) An array of AWS regions to exclude from metrics collection.
- `extended_resource_collection_enabled` (String) Whether Datadog collects additional attributes and configuration information about the resources in your AWS account. Required for `cspm_resource_collection_enabled`.
- `filter_tags` (List of String) Array of EC2 tags (in the form `key:value`) defines a filter that Datadog uses when collecting metrics from EC2. Wildcards, such as `?` (for single characters) and `*` (for multiple characters) can also be used. Only hosts that match one of the defined tags will be imported into Datadog. The rest will be ignored. Host matching a given tag can also be excluded by adding `!` before the tag. e.x. `env:production,instance-type:c1.*,!region:us-east-1`.
- `host_tags` (List of String) Array of tags (in the form `key:value`) to add to all hosts and metrics reporting through this integration.
- `metrics_collection_enabled` (String) Whether Datadog collects metrics for this AWS account.
- `resource_collection_enabled` (String, Deprecated) Whether Datadog collects a standard set of resources from your AWS account. **Deprecated.** Deprecated in favor of `extended_resource_collection_enabled`.
- `role_name` (String) Your Datadog role delegation name.
- `secret_access_key` (String, Sensitive) Your AWS secret access key. Only required if your AWS account is a GovCloud or China account.

### Read-Only

- `external_id` (String) AWS External ID. **NOTE** This provider will not be able to detect changes made to the `external_id` field from outside Terraform.
- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# Amazon Web Services integrations can be imported using their account ID and role name separated with a colon (:), while the external_id should be passed by setting an environment variable called EXTERNAL_ID
EXTERNAL_ID=${external_id} terraform import datadog_integration_aws.test ${account_id}:${role_name}
```
