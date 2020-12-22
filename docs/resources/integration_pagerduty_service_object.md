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

- **service_key** (String, Required)
- **service_name** (String, Required)

### Optional

- **id** (String, Optional) The ID of this resource.


