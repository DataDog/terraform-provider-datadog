---
layout: "datadog"
page_title: "Datadog: datadog_integration_pagerduty"
sidebar_current: "docs-datadog-resource-integration_pagerduty"
description: |-
  Provides a Datadog - PagerDuty integration resource. This can be used to create and manage the integration.
---

# datadog_integration_pagerduty

Provides a Datadog - PagerDuty resource. This can be used to create and manage Datadog - PagerDuty integration. 

## Example Usage

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

## Argument Reference

The following arguments are supported:

* `services` - (Optional) Array of PagerDuty service objects.
  * `service_name` - (Required) A name for your service to use with Datadog.
  * `service_key` - (Required) The service's integration key from PagerDuty.
* `schedules` - (Optional)  Array of your schedule URLs.
* `subdomain` - (Required) Your PagerDuty account’s personalized subdomain name.
* `api_token` - (Optional) Your PagerDuty API token.


# datadog_integration_pagerduty_service

Provides a resource which connects a single PagerDuty service to Datadog.

## Example Usage

```
data "pagerduty_vendor" "datadog" {
  name = "Datadog"
}

resource "pagerduty_service" "app" {
  name                    = "example-for-datadog"
  description             = "Example pagerduty service"
  acknowledgement_timeout = 1800
  auto_resolve_timeout    = "null"
  escalation_policy       = "default"
  alert_creation          = "create_alerts_and_incidents"
}

resource "pagerduty_service_integration" "datadog" {
  name    = "${data.pagerduty_vendor.datadog.name} example-for-datadog"
  service = "${pagerduty_service.app.id}"
  vendor  = "${data.pagerduty_vendor.datadog.id}"
}

resource "datadog_integration_pagerduty_service" "datadog_service" {
  service_name = "DatadogPdService"
  service_key  = "${pagerduty_service_integration.datadog.integration_key}"
}

resource "datadog_monitor" "datadog_service" {
  name                = "Monitor for the Datadog service"
  type                = "metric alert"
  message             = "Monitor triggered. Notify: ${datadog_integration_pagerduty_service.datadog_service.notify_handle}"
  query               = "avg(last_1h):avg:aws.ec2.cpuutilization{*} > 50"
  notify_no_data      = false
  require_full_window = false
  tags                = ["terraform:true"]
}
```

## Argument Reference

The following arguments are supported:

* `service_name`  - (Required) A name for your service to use with Datadog.
* `service_key`   - (Required) The service's integration key from PagerDuty.
* `notify_handle` - (Computed) Handle which can be used in monitors to link it with this PagerDuty service.

### See also
* [PagerDuty Integration Guide](https://www.pagerduty.com/docs/guides/datadog-integration-guide/)
* [Datadog API Reference > Integrations > PagerDuty](https://docs.datadoghq.com/api/?lang=bash#pagerduty)
