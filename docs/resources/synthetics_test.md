---
page_title: "datadog_synthetics_test"
---

# datadog_synthetics_test Resource

Provides a Datadog synthetics test resource. This can be used to create and manage Datadog synthetics test.

## Example Usage (Synthetics API test)

Create a new Datadog Synthetics API/HTTP test on https://www.example.org

```hcl
resource "datadog_synthetics_test" "test_api" {
  type = "api"
  subtype = "http"
  request = {
    method = "GET"
    url = "https://www.example.org"
  }
  request_headers = {
    Content-Type = "application/json"
    Authentication = "Token: 1234566789"
  }
  assertion {
      type = "statusCode"
      operator = "is"
      target = "200"
  }
  locations = [ "aws:eu-central-1" ]
  options_list {
    tick_every = 900

    retry {
      count = 2
      interval = 300
    }

    monitor_options {
      renotify_interval = 100
    }
  }
  name = "An API test on example.org"
  message = "Notify @pagerduty"
  tags = ["foo:bar", "foo", "env:test"]

  status = "live"
}
```

## Example Usage (Synthetics SSL test)

Create a new Datadog Synthetics API/SSL test on example.org

```hcl
resource "datadog_synthetics_test" "test_ssl" {
  type = "api"
  subtype = "ssl"
  request = {
    host = "example.org"
    port = 443
  }
  assertion {
      type = "certificate"
      operator = "isInMoreThan"
      target = 30
  }
  locations = [ "aws:eu-central-1" ]
  options_list {
    tick_every = 900
    accept_self_signed = true
  }
  name = "An API test on example.org"
  message = "Notify @pagerduty"
  tags = ["foo:bar", "foo", "env:test"]

  status = "live"
}
```

## Example Usage (Synthetics TCP test)

Create a new Datadog Synthetics API/TCP test on example.org

```hcl
resource "datadog_synthetics_test" "test_tcp" {
  type = "api"
  subtype = "tcp"
  request = {
    host = "example.org"
    port = 443
  }
  assertion {
      type = "responseTime"
      operator = "lessThan"
      target = 2000
  }
  locations = [ "aws:eu-central-1" ]
  options_list {
    tick_every = 900
  }
  name = "An API test on example.org"
  message = "Notify @pagerduty"
  tags = ["foo:bar", "foo", "env:test"]

  status = "live"
}
```

## Example Usage (Synthetics DNS test)

Create a new Datadog Synthetics API/DNS test on example.org

```hcl
resource "datadog_synthetics_test" "test_dns" {
  type = "api"
  subtype = "dns"
  request = {
    host = "example.org"
  }
  assertion {
    type = "recordSome"
    operator = "is"
    property = "A"
    target = "0.0.0.0"
  }
  locations = [ "aws:eu-central-1" ]
  options_list {
    tick_every = 900
  }
  name = "An API test on example.org"
  message = "Notify @pagerduty"
  tags = ["foo:bar", "foo", "env:test"]

  status = "live"
}
```

## Example Usage (Synthetics Browser test)

Support for Synthetics Browser test steps is limited (see [below](#synthetics-browser-test))

```hcl
# Create a new Datadog Synthetics Browser test starting on https://www.example.org
resource "datadog_synthetics_test" "test_browser" {
  type = "browser"

  request = {
    method = "GET"
    url    = "https://app.datadoghq.com"
  }

  device_ids = ["laptop_large"]
  locations  = ["aws:eu-central-1"]

  options_list {
    tick_every = 3600
  }

  name = "A Browser test on example.org"
  message = "Notify @qa"
  tags    = []

  status = "paused"

  step {
    name = "Check current url"
    type = "assertCurrentUrl"
    params = jsonencode({
        "check": "contains",
        "value": "datadoghq"
    })
  }

  browser_variable {
    type    = "text"
    name    = "MY_PATTERN_VAR"
    pattern = "{{numeric(3)}}"
    example = "597"
  }

  browser_variable {
    type    = "email"
    name    = "MY_EMAIL_VAR"
    pattern = "jd8-afe-ydv.{{ numeric(10) }}@synthetics.dtdg.co"
    example = "jd8-afe-ydv.4546132139@synthetics.dtdg.co"
  }

  browser_variable {
    type = "global"
    name = "MY_GLOBAL_VAR"
    id   = "76636cd1-82e2-4aeb-9cfe-51366a8198a2"
  }
}
```

## Argument Reference

The following arguments are supported:

-   `type`: (Required) Synthetics test type (api or browser)
-   `subtype`: (Optional) For type=api, http, ssl, tcp or dns (Default = http)
-   `name`: (Required) Name of Datadog synthetics test
-   `message`: (Required) A message to include with notifications for this synthetics test. Email notifications can be sent to specific users by using the same '@username' notation as events.
-   `tags`: (Optional) A list of tags to associate with your synthetics test. This can help you categorize and filter tests in the manage synthetics page of the UI. Default is an empty list ([]).
-   `request`: (Required) if type=api and subtype=http
    -   `method`: (Optional) For type=api and subtype=http, one of DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT
    -   `url`: (Required) Any url
    -   `timeout`: (Optional) For type=api, any value between 0 and 60 (Default = 60)
    -   `body`: (Optional) Request body
-   `request`: (Required) if type=api and subtype=ssl or subtype=tcp or subtype=dns
    -   `host`: (Required) host name
    -   `port`: (Required) port number
    -   `timeout`: (Optional) For type=api, any value between 0 and 60 (Default = 60)
    -   `dns_server`: (Optional) For subtype=dns, DNS server to use
-   `request`: (Required) if type=browser
    -   `method`: (Required) no-op, use GET
    -   `url`: (Required) Any url
-   `request_headers`: (Optional) Header name and value map
-   `request_query`: (Optional) Query arguments name and value map
-   `request_basicauth`: (Optional) Array of 1 item containing HTTP basic authentication credentials
    -   `username`: (Required) Username for authentication
    -   `password`: (Required) Password for authentication
-   `request_client_certificate`: (Optional) Client certificate to use when performing the test request
    -   `cert`
        -   `content`: (Required) Content of the client certificate
        -   `filename`: (Optional) Filename for the certificate
    -   `key`
        -   `content`: (Required) Content of the certificate key
        -   `filename`: (Optional) Filename for the certificate key
-   `assertion`: (Required) Array of 1 to 10 items, only some combinations of type/operator are valid (please refer to Datadog documentation).
    -   `type`: (Required) body, header, responseTime, statusCode
    -   `operator`: (Required) Please refer to [Datadog documentation](https://docs.datadoghq.com/synthetics/api_test/#validation) as operator depend on assertion type
    -   `target`: (Optional) Expected value, please refer to [Datadog documentation](https://docs.datadoghq.com/synthetics/api_test/#validation) as target depend on assertion type
    -   `targetjsonpath`: (Optional) Expected structure if `operator` is `validatesJSONPath`
        -   `operator`: (Required) The specific operator to use on the path
        -   `targetvalue`: (Required) Expected matching value
        -   `jsonpath`: (Required) The JSON path to assert
    -   `property`: (Optional) if assertion type is "header", this is a the header name
-   `options`: (Required) **Deprecated**
    -   `tick_every`: (Required) How often the test should run (in seconds). Current possible values are 900, 1800, 3600, 21600, 43200, 86400, 604800 plus 60 if type=api or 300 if type=browser
    -   `follow_redirects`: (Optional) For type=api, true or false
    -   `min_failure_duration`: (Optional) How long the test should be in failure before alerting (integer, number of seconds, max 7200). Default is 0.
    -   `min_location_failed`: (Optional) Minimum number of locations in failure required to trigger an alert.
    -   `accept_self_signed`: (Optional) For type=ssl, true or false
    -   `allow_insecure`: (Optional) For type=api, true or false. Allow your HTTP test go on with connection even if there is an error when validating the certificate.
    -   `retry_count`: (Optional) Number of retries needed to consider a location as failed before sending a notification alert.
    -   `retry_interval`: (Optional) Interval between a failed test and the next retry in milliseconds.
-   `options_list`: (Optional)
    -   `tick_every`: (Optional) How often the test should run (in seconds). Current possible values are 900, 1800, 3600, 21600, 43200, 86400, 604800 plus 60 if type=api or 300 if type=browser
    -   `follow_redirects`: (Optional) For type=api, true or false
    -   `min_failure_duration`: (Optional) How long the test should be in failure before alerting (integer, number of seconds, max 7200). Default is 0.
    -   `min_location_failed`: (Optional) Threshold below which a synthetics test is allowed to fail before sending notifications. Default is 1.
    -   `accept_self_signed`: (Optional) For type=ssl, true or false
    -   `allow_insecure`: (Optional) For type=api, true or false. Allow your HTTP test go on with connection even if there is an error when validating the certificate.
    -   `retry`: (Optional)
        -   `count`: (Optional) Number of retries needed to consider a location as failed before sending a notification alert.
        -   `interval`: (Optional) Interval between a failed test and the next retry in milliseconds.
    -   `monitor_options`: (Optional)
        -   `renotification_interval`: (Optional) Specify a renotification frequency.
-   `locations`: (Required) Please refer to [Datadog documentation](https://docs.datadoghq.com/synthetics/api_test/#request) for available locations (e.g. "aws:eu-central-1")
-   `device_ids`: (Optional) "laptop_large", "tablet" or "mobile_small" (only available if type=browser)
-   `status`: (Required) "live", "paused"
-   `step`: (Optional) Steps for browser tests.
    -   `name`: (Required) Name of the step.
    -   `type`: (Required) Type of step. Please refer to [Datadog documentation](https://docs.datadoghq.com/api/v1/synthetics/#create-a-test) for the complete list of step type available.
    -   `params`: (Required) Parameters for the step as JSON string.
    -   `allow_failure`: (Optional) Determines if the step should be allowed to fail.
    -   `timeout`: (Optional) Used to override the default timeout of a step.
-   `variable`: (Optional) **Deprecated** Array of variables used for the test.
    -   `type`: (Required) Type of browser test variable. Allowed enum values: "element","email","global","text"
    -   `name`: (Required) Name of the variable.
    -   `example`: (Optional) Example for the variable.
    -   `id`: (Optional) ID of the global variable to use. This is actually only used (and required) in the case of using a variable of type "global".
    -   `pattern`: (Optional) Pattern of the variable.
-   `browser_variable`: (Optional) Only for browser tests. Array of variables used for the test.
    -   `type`: (Required) Type of browser test variable. Allowed enum values: "element","email","global","text"
    -   `name`: (Required) Name of the variable.
    -   `example`: (Optional) Example for the variable.
    -   `id`: (Optional) ID of the global variable to use. This is actually only used (and required) in the case of using a variable of type "global".
    -   `pattern`: (Optional) Pattern of the variable.
-   `config_variable`: (Optional) Only for api tests. Array of variables used for the test configuration.
    -   `type`: (Required) Type of test variable. Allowed enum values: "text".
    -   `name`: (Required) Name of the variable.
    -   `example`: (Optional) Example for the variable.
    -   `pattern`: (Optional) Pattern of the variable.

## Attributes Reference

The following attributes are exported:

-   `id`: ID (public_id) of the Datadog synthetics test
-   `monitor_id`: ID of the monitor associated with the Datadog synthetics test

## Import

Synthetics tests can be imported using their public string ID, e.g.

```
$ terraform import datadog_synthetics_test.fizz abc-123-xyz
```

## Synthetics Browser test

Support for Synthetics Browser test is limited when creating steps. Some steps types (like steps involving elements) cannot be created, but they can be imported.

## Assertion format

The resource was changed to have assertions be a list of `assertion` blocks instead of single `assertions` array, to support the JSON path operations. We'll remove `assertions` support in the future: to migrate, rename your attribute to `assertion` and turn array elements into independent blocks. For example:

```hcl
resource "datadog_synthetics_test" "test_api" {
  assertions = [
    {
      type = "statusCode"
      operator = "is"
      target = "200"
    },
    {
      type = "responseTime"
      operator = "lessThan"
      target = "1000"
    }
  ]
}
```

turns into:

```hcl
resource "datadog_synthetics_test" "test_api" {
  assertion {
      type = "statusCode"
      operator = "is"
      target = "200"
  }
  assertion {
      type = "responseTime"
      operator = "lessThan"
      target = "1000"
  }
}
```
