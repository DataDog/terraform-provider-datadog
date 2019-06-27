---
layout: "datadog"
page_title: "Datadog: datadog_integration_pagerduty_service_object"
sidebar_current: "docs-datadog-resource-integration_pagerduty_service_object"
description: |-
  Provides a Datadog - PagerDuty integration resource. This can be used to create and manage the integration.
---

# datadog_integration_pagerduty_service_object

Provides access to individual Service Objects of Datadog - PagerDuty integrations. Note that the Datadog - PagerDuty integration must be activated (either manually in the Datadog UI or by using [datadog_integration_pagerduty](/docs/providers/datadog/r/integration_pagerduty.html)) in order for this resource to be usable.

## Example Usage

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

## Argument Reference

The following arguments are supported:

* `service_name` - (Required) Your Service name in PagerDuty.
* `service_key` - (Required) Your Service name associated service key in PagerDuty. Note: Since the Datadog API never returns service keys, it is impossible to detect [drifts](https://www.hashicorp.com/blog/detecting-and-managing-drift-with-terraform). The best way to solve a drift is to manually mark the Service Object resource with [terraform taint](https://www.terraform.io/docs/commands/taint.html) to have it destroyed and recreated.