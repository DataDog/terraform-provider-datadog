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
  * `service_name` - (Required) Your Service name in PagerDuty.
  * `service_key` - (Required) Your Service name associated service key in Pagerduty.
* `schedules` - (Optional)  Array of your schedule URLs.
* `subdomain` - (Required) Your PagerDuty account’s personalized subdomain name.
* `api_token` - (Optional) Your PagerDuty API token.

### See also
* [PagerDuty Integration Guide](https://www.pagerduty.com/docs/guides/datadog-integration-guide/)
* [Datadog API Reference > Integrations > PagerDuty](https://docs.datadoghq.com/api/?lang=bash#pagerduty)
