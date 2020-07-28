---
page_title: "datadog_integration_pagerduty"
---

# datadog_integration_pagerduty Resource

Provides a Datadog - PagerDuty resource. This can be used to create and manage Datadog - PagerDuty integration. This resource is deprecated and should only be used for legacy purposes.

## Example Usage

*Note:* Until terraform-provider-datadog version 2.1.0, service objects under the `services` key were specified inside the `datadog_integration_pagerduty` resource. This was incompatible with multi-configuration-file setups, where users wanted to have individual service objects controlled from different Terraform configuration files. The recommended approach now is specifying service objects as individual resources using [datadog_integration_pagerduty_service_object](/docs/providers/datadog/r/integration_pagerduty_service_object.html) and adding `individual_services = true` to the `datadog_integration_pagerduty` object.

### Services as Individual Resources

```
resource "datadog_integration_pagerduty" "pd" {
  individual_services = true
  schedules = [
    "https://ddog.pagerduty.com/schedules/X123VF",
    "https://ddog.pagerduty.com/schedules/X321XX"
    ]
  subdomain = "ddog"
  api_token = "38457822378273432587234242874"
}

resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  # when creating the integration object for the first time, the service
  # objects have to be created *after* the integration
  depends_on = ["datadog_integration_pagerduty.pd"]
  service_name = "testing_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  depends_on = ["datadog_integration_pagerduty.pd"]
  service_name = "testing_bar"
  service_key  = "54321098765432109876"
}
```

### Inline Services

With Terraform < 0.12.0 (terraform-provider-datadog < 1.9.0):

```
# Create a new Datadog - PagerDuty integration
resource "datadog_integration_pagerduty" "pd" {
  services = [
    {
      service_name = "testing_foo"
      service_key  = "9876543210123456789"
    },
    {
      service_name = "testing_bar"
      service_key = "54321098765432109876"
    }
  ]
  schedules = [
    "https://ddog.pagerduty.com/schedules/X123VF",
    "https://ddog.pagerduty.com/schedules/X321XX"
    ]
  subdomain = "ddog"
  api_token = "38457822378273432587234242874"
}
```

With Terraform >= 0.12.0 (terraform-provider-datadog >= 1.9.0):

```
locals {
  pd_services = {
    testing_foo = "9876543210123456789"
    testing_bar = "54321098765432109876"
  }
}
# Create a new Datadog - PagerDuty integration
resource "datadog_integration_pagerduty" "pd" {
  dynamic "services" {
    for_each = local.pd_services
    content {
      service_name = services.key
      service_key = services.value
    }
  }
  schedules = [
    "https://ddog.pagerduty.com/schedules/X123VF",
    "https://ddog.pagerduty.com/schedules/X321XX"
    ]
  subdomain = "ddog"
  api_token = "38457822378273432587234242874"
}
```

### Migrating from Inline Services to Individual Resources

Migrating from usage of inline services to individual resources is very simple. The following example shows how to convert an existing inline services configuration to configuration using individual resources. Doing analogous change and running `terraform apply` after every step is all that's necessary to migrate.

```
# First step - this is what the configuration looked like initially

locals {
  pd_services = {
    testing_foo = "9876543210123456789"
    testing_bar = "54321098765432109876"
  }
}
# Create a new Datadog - PagerDuty integration
resource "datadog_integration_pagerduty" "pd" {
  dynamic "services" {
    for_each = local.pd_services
    content {
      service_name = services.key
      service_key = services.value
    }
  }
  schedules = [
    "https://ddog.pagerduty.com/schedules/X123VF",
    "https://ddog.pagerduty.com/schedules/X321XX"
    ]
  subdomain = "ddog"
  api_token = "38457822378273432587234242874"
}
```

```
# Second step - this will remove the inline-defined service objects
# Note that during this step, `individual_services` must not be defined
resource "datadog_integration_pagerduty" "pd" {
  # `services` was removed
  schedules = [
    "https://ddog.pagerduty.com/schedules/X123VF",
    "https://ddog.pagerduty.com/schedules/X321XX"
    ]
  subdomain = "ddog"
  api_token = "38457822378273432587234242874"
}
```

```
# Third step - this will reintroduce the service objects as individual resources

resource "datadog_integration_pagerduty" "pd" {
  # `individual_services = true` was added
  individual_services = true
  schedules = [
    "https://ddog.pagerduty.com/schedules/X123VF",
    "https://ddog.pagerduty.com/schedules/X321XX"
    ]
  subdomain = "ddog"
  api_token = "38457822378273432587234242874"
}

resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  depends_on = ["datadog_integration_pagerduty.pd"]
  service_name = "testing_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  depends_on = ["datadog_integration_pagerduty.pd"]
  service_name = "testing_bar"
  service_key  = "54321098765432109876"
}
```

## Argument Reference

The following arguments are supported:

* `individual_services` - (Optional) Boolean to specify whether or not individual service objects specified by [datadog_integration_pagerduty_service_object](/docs/providers/datadog/r/integration_pagerduty_service_object.html) resource are to be used. Mutually exclusive with `services` key.
* `services` - (Optional) Array of PagerDuty service objects. **Deprecated** The `services` list is now deprecated in favour of [datadog_integration_pagerduty_service_object](/docs/providers/datadog/r/integration_pagerduty_service_object.html) resource. Note that `individual_services` must be set to `true` to ignore the `service` attribute and use individual services properly.
  * `service_name` - (Required) Your Service name in PagerDuty.
  * `service_key` - (Required) Your Service name associated service key in Pagerduty.
* `schedules` - (Optional)  Array of your schedule URLs.
* `subdomain` - (Required) Your PagerDuty account’s personalized subdomain name.
* `api_token` - (Optional) Your PagerDuty API token.

### See also
* [PagerDuty Integration Guide](https://www.pagerduty.com/docs/guides/datadog-integration-guide/)
* [Datadog API Reference > Integrations > PagerDuty](https://docs.datadoghq.com/api/v1/pagerduty-integration/)
