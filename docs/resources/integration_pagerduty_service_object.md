---
page_title: "datadog_integration_pagerduty_service_object Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides access to individual Service Objects of Datadog - PagerDuty integrations. Note that the Datadog - PagerDuty integration must be activated in the Datadog UI in order for this resource to be usable.
---

# Resource `datadog_integration_pagerduty_service_object`

Provides access to individual Service Objects of Datadog - PagerDuty integrations. Note that the Datadog - PagerDuty integration must be activated in the Datadog UI in order for this resource to be usable.

## Example Usage

```terraform
resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  service_name = "testing_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  service_name = "testing_bar"
  service_key  = "54321098765432109876"
}
```

## Schema

### Required

- **service_key** (String, Sensitive) Your Service name associated service key in PagerDuty. Note: Since the Datadog API never returns service keys, it is impossible to detect [drifts](https://www.hashicorp.com/blog/detecting-and-managing-drift-with-terraform?_ga=2.15990198.1091155358.1609189257-888022054.1605547463). The best way to solve a drift is to manually mark the Service Object resource with [terraform taint](https://www.terraform.io/docs/commands/taint.html?_ga=2.15990198.1091155358.1609189257-888022054.1605547463) to have it destroyed and recreated.
- **service_name** (String) Your Service name in PagerDuty.

### Optional

- **id** (String) The ID of this resource.


