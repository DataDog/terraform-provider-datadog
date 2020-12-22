---
page_title: "datadog_integration_azure Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog - Microsoft Azure integration resource. This can be used to create and manage the integrations.
---

# Resource `datadog_integration_azure`

Provides a Datadog - Microsoft Azure integration resource. This can be used to create and manage the integrations.

## Example Usage

```terraform
# Create a new Datadog - Microsoft Azure integration
resource "datadog_integration_azure" "sandbox" {
    tenant_name = "<azure_tenant_name>"
    client_id = "<azure_client_id>"
    client_secret = "<azure_client_secret_key>"
    host_filters = "examplefilter:true,example:true"
}
```

## Schema

### Required

- **client_id** (String, Required)
- **client_secret** (String, Required)
- **tenant_name** (String, Required)

### Optional

- **host_filters** (String, Optional)
- **id** (String, Optional) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# Microsoft Azure integrations can be imported using their `tenant name` and `client` id separated with a colon (`:`).

terraform import datadog_integration_azure.sandbox ${tenant_name}:${client_id}
```
