---
page_title: "datadog_integration_pagerduty_service_object"
---

# datadog_integration_pagerduty_service_object Resource

Provides access to individual Service Objects of Datadog - PagerDuty integrations. Note that the Datadog - PagerDuty integration must be activated in the Datadog UI in order for this resource to be usable.

## Example Usage

```
resource "datadog_integration_pagerduty_service_object" "testing_foo" {
  service_name = "testing_foo"
  service_key  = "9876543210123456789"
}

resource "datadog_integration_pagerduty_service_object" "testing_bar" {
  service_name = "testing_bar"
  service_key  = "54321098765432109876"
}
```

## Argument Reference

The following arguments are supported:

- `service_name`: (Required) Your Service name in PagerDuty.
- `service_key`: (Required) Your Service name associated service key in PagerDuty. Note: Since the Datadog API never returns service keys, it is impossible to detect [drifts](https://www.hashicorp.com/blog/detecting-and-managing-drift-with-terraform). The best way to solve a drift is to manually mark the Service Object resource with [terraform taint](https://www.terraform.io/docs/commands/taint.html) to have it destroyed and recreated.
