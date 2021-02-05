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
  tenant_name   = "<azure_tenant_name>"
  client_id     = "<azure_client_id>"
  client_secret = "<azure_client_secret_key>"
  host_filters  = "examplefilter:true,example:true"
}
```

## Schema

### Required

- **client_id** (String, Required) Your Azure web application ID.
- **client_secret** (String, Required) (Required for Initial Creation) Your Azure web application secret key.
- **tenant_name** (String, Required) Your Azure Active Directory ID.

### Optional

- **host_filters** (String, Optional) String of host tag(s) (in the form `key:value,key:value`) defines a filter that Datadog will use when collecting metrics from Azure. Limit the Azure instances that are pulled into Datadog by using tags. Only hosts that match one of the defined tags are imported into Datadog. e.x. `env:production,deploymentgroup:red`
- **id** (String, Optional) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# Microsoft Azure integrations can be imported using their `tenant name` and `client` id separated with a colon (`:`).
terraform import datadog_integration_azure.sandbox ${tenant_name}:${client_id}
```
