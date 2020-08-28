---
page_title: "datadog_integration_azure"
---

# datadog_integration_azure Resource

Provides a Datadog - Microsoft Azure integration resource. This can be used to create and manage the integrations.

## Example Usage

```hcl
# Create a new Datadog - Microsoft Azure integration
resource "datadog_integration_azure" "sandbox" {
    tenant_name = "<azure_tenant_name>"
    client_id = "<azure_client_id>"
    client_secret = "<azure_client_secret_key>"
    host_filters = "examplefilter:true,example:true"
}
```

## Argument Reference

The following arguments are supported:

* `tenant_name`: (Required) Your Azure Active Directory ID.
* `client_id`: (Required) Your Azure web application ID.
* `client_secret`: (Required for Initial Creation) Your Azure web application secret key.
* `host_filters`: (Optional) String of host tag(s) (in the form `key:value,key:value`) defines a filter that Datadog will use when collecting metrics from Azure.

  Limit the Azure instances that are pulled into Datadog by using tags. Only hosts that match one of the defined tags are imported into Datadog.

  e.x. `env:production,deploymentgroup:red`

### See also
* [Datadog API Reference > Integrations > Azure](https://https://docs.datadoghq.com/integrations/azure/)

## Import

Microsoft Azure integrations can be imported using their `tenant name` and `client id` separated with a colon (`:`).

```
$ terraform import datadog_integration_azure.sandbox ${tenant_name}:${client_id}
```
