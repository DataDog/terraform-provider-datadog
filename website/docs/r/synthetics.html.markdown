---
layout: "datadog"
page_title: "Datadog: datadog_synthetics_test"
sidebar_current: "docs-datadog-resource-synthetics_test"
description: |-
  Provides a Datadog synthetics resource. This can be used to create and manage synthetics.
---

# datadog_synthetics_test

Provides a Datadog synthetics test resource. This can be used to create and manage Datadog synthetics test.

## Example Usage

```hcl
# Create a new Datadog Synthetics API test on https://www.example.org
resource "datadog_synthetics_test" "foo" {
  type = "api"
  request {
    method = "GET"
    url = "https://www.example.org"
  }
  request_headers {
    "Content-Type" = "application/json"
    "Authentication" = "Token: 1234566789"
  }
  assertions = [
    {
      type = "statusCode"
      operator = "is"
      target = "200"
    }
  ]
  locations = [ "aws:eu-central-1" ]
  options {
    tick_every = 900
  }
  name = "An API test on example.org"
  message = "Notify @pagerduty"
  tags = ["foo:bar", "foo", "env:test"]

  status = "live"
}

# Create a new Datadog Synthetics Browser test starting on https://www.example.org
resource "datadog_synthetics_test" "bar" {
  type = "browser"

  request {
    method = "GET"
    url    = "https://app.datadoghq.com"
  }

  device_ids = ["laptop_large"]
  locations  = ["aws:eu-central-1"]

  options {
    tick_every = 3600
  }

  name = "A Browser test on example.org"
  message = "Notify @qa"
  tags    = []

  status = "paused"
}
```

## Argument Reference

The following arguments are supported:

- `type` - (Required) Synthetics test type (api or browser)
- `name` - (Required) Name of Datadog synthetics test
- `message` - (Required) A message to include with notifications for this synthetics test.
  Email notifications can be sent to specific users by using the same '@username' notation as events.
- `tags` - (Required) A list of tags to associate with your synthetics test. This can help you categorize and filter tests in the manage synthetics page of the UI.
- `request` - (Required)
  - `method` - (Required) DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT (noop if type=browser)
  - `url` - (Required) Any url
  - `timeout` - (Optional) For type=api, any value between 0 and 60 (Default = 60). No-op for type=browser
  - `body` - (Optional) Request body
- `request_headers` - (Optional) Header name and value map
- `assertions` - (Required) Array of 1 to 10 items, only some combinations of type/operator are valid (please refer to Datadog documentation)
  - `type` - (Required) body, header, responseTime, statusCode
  - `operator` - (Required) "contains", "doesNotContain", "is", "isNot", "matches", "doesNotMatch", "validates"
  - `target` - (Required) Expected string
  - `property` - (Optional)
- `options` - (Required)
  - `tick_every` - (Required) 900, 1800, 3600, 21600, 43200, 86400, 604800 plus 60 if type=api or 300 if type=browser
  - `follow_redirects` - (Optional) true or false
  - `min_failure_duration` - (Optional) Grace period during which a synthetics test is allowed to fail before sending notifications
  - `min_location_failed` - (Optional) Threshold below which a synthetics test is allowed to fail before sending notifications
- `locations` - (Required) Please refer to Datadog documentation for available locations (e.g. "aws:eu-central-1")
- `device_ids` - (Optional) "laptop_large", "tablet" or "mobile_small" (only available if type=browser)
- `status` - (Required) "live", "paused"

## Attributes Reference

The following attributes are exported:

- `public_id` - ID of the Datadog synthetics test

## Import

Synthetics tests can be imported using their public string ID, e.g.

```
$ terraform import datadog_synthetics_test.fizz abc-123-xyz
```
