---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_integration_azure Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog - Microsoft Azure integration resource. This can be used to create and manage the integrations.
---

# datadog_integration_azure (Resource)

Provides a Datadog - Microsoft Azure integration resource. This can be used to create and manage the integrations.

## Example Usage

```terraform
# Create a new Datadog - Microsoft Azure integration
resource "datadog_integration_azure" "sandbox" {
  tenant_name              = "<azure_tenant_name>"
  client_id                = "<azure_client_id>"
  client_secret            = "<azure_client_secret_key>"
  host_filters             = "examplefilter:true,example:true"
  app_service_plan_filters = "examplefilter:true,example:another"
  container_app_filters    = "examplefilter:true,example:one_more"
  automute                 = true
  cspm_enabled             = true
  custom_metrics_enabled   = false
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `client_id` (String) Your Azure web application ID.
- `client_secret` (String, Sensitive) (Required for Initial Creation) Your Azure web application secret key.
- `tenant_name` (String) Your Azure Active Directory ID.

### Optional

- `app_service_plan_filters` (String) This comma-separated list of tags (in the form `key:value,key:value`) defines a filter that Datadog uses when collecting metrics from Azure App Service Plans. Only App Service Plans that match one of the defined tags are imported into Datadog. The rest, including the apps and functions running on them, are ignored. This also filters the metrics for any App or Function running on the App Service Plan(s). Defaults to `""`.
- `automute` (Boolean) Silence monitors for expected Azure VM shutdowns. Defaults to `false`.
- `container_app_filters` (String) This comma-separated list of tags (in the form `key:value,key:value`) defines a filter that Datadog uses when collecting metrics from Azure Container Apps. Only Container Apps that match one of the defined tags are imported into Datadog. Defaults to `""`.
- `cspm_enabled` (Boolean) When enabled, Datadog’s Cloud Security Management product scans resource configurations monitored by this app registration.
Note: This requires `resource_collection_enabled` to be set to true. Defaults to `false`.
- `custom_metrics_enabled` (Boolean) Enable custom metrics for your organization. Defaults to `false`.
- `host_filters` (String) String of host tag(s) (in the form `key:value,key:value`) defines a filter that Datadog will use when collecting metrics from Azure. Limit the Azure instances that are pulled into Datadog by using tags. Only hosts that match one of the defined tags are imported into Datadog. e.x. `env:production,deploymentgroup:red` Defaults to `""`.
- `metrics_enabled` (Boolean) Enable Azure metrics for your organization. Defaults to `true`.
- `metrics_enabled_default` (Boolean) Enable Azure metrics for your organization for resource providers where no resource provider config is specified. Defaults to `true`.
- `resource_collection_enabled` (Boolean) When enabled, Datadog collects metadata and configuration info from cloud resources (such as compute instances, databases, and load balancers) monitored by this app registration.
- `resource_provider_configs` (List of Object) Configuration settings applied to resources from the specified Azure resource providers. (see [below for nested schema](#nestedatt--resource_provider_configs))
- `usage_metrics_enabled` (Boolean) Enable azure.usage metrics for your organization. Defaults to `true`.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedatt--resource_provider_configs"></a>
### Nested Schema for `resource_provider_configs`

Optional:

- `metrics_enabled` (Boolean)
- `namespace` (String)

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# Microsoft Azure integrations can be imported using their `tenant name` and `client` id separated with a colon (`:`).
# The client_secret should be passed by setting the environment variable CLIENT_SECRET
terraform import datadog_integration_azure.sandbox ${tenant_name}:${client_id}
```
