---
page_title: "datadog_integration_aws Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  
---

# Resource `datadog_integration_aws`



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

- **account_id** (String, Required)
- **role_name** (String, Required)

### Optional

- **account_specific_namespace_rules** (Map of Boolean, Optional)
- **excluded_regions** (List of String, Optional)
- **filter_tags** (List of String, Optional)
- **host_tags** (List of String, Optional)
- **id** (String, Optional) The ID of this resource.

### Read-only

- **external_id** (String, Read-only)

## Import

Import is supported using the following syntax:

```shell
# Amazon Web Services integrations can be imported using their account ID and role name separated with a colon (:), while the external_id should be passed by setting an environment variable called EXTERNAL_ID

EXTERNAL_ID=${external_id} terraform import datadog_integration_aws.test ${account_id}:${role_name}
```
