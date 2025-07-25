---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_app_builder_app Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  This data source retrieves the definition of an existing Datadog App from App Builder for use in other resources, such as embedding Apps in Dashboards. This data source requires a registered application key https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/app_key_registration.
---

# datadog_app_builder_app (Data Source)

This data source retrieves the definition of an existing Datadog App from App Builder for use in other resources, such as embedding Apps in Dashboards. This data source requires a [registered application key](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/app_key_registration).

## Example Usage

```terraform
data "datadog_app_builder_app" "my_app" {
  id = "11111111-2222-3333-4444-555555555555"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) ID for the App.

### Read-Only

- `action_query_names_to_connection_ids` (Map of String) A map of the App's Action Query Names to Action Connection IDs.
- `app_json` (String) The JSON representation of the App.
- `description` (String) The human-readable description of the App.
- `name` (String) The name of the App.
- `published` (Boolean) Whether the app is published or unpublished. Published apps are available to other users. To ensure the app is accessible to the correct users, you also need to set a [Restriction Policy](https://docs.datadoghq.com/api/latest/restriction-policies/) on the app if a policy does not yet exist.
- `root_instance_name` (String) The name of the root component of the app. This is a grid component that contains all other components.
