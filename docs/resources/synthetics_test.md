---
page_title: "datadog_synthetics_test Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog synthetics test resource. This can be used to create and manage Datadog synthetics test.
---

# Resource `datadog_synthetics_test`

Provides a Datadog synthetics test resource. This can be used to create and manage Datadog synthetics test.

## Example Usage

```terraform
# Example Usage (Synthetics API test)
# Create a new Datadog Synthetics API/HTTP test on https://www.example.org
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


# Example Usage (Synthetics SSL test)
# Create a new Datadog Synthetics API/SSL test on example.org
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


# Example Usage (Synthetics TCP test)
# Create a new Datadog Synthetics API/TCP test on example.org
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


# Example Usage (Synthetics DNS test)
# Create a new Datadog Synthetics API/DNS test on example.org
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


# Example Usage (Synthetics Browser test)
# Support for Synthetics Browser test steps is limited (see below)
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

  variable {
    type    = "text"
    name    = "MY_PATTERN_VAR"
    pattern = "{{numeric(3)}}"
    example = "597"
  }

  variable {
    type    = "email"
    name    = "MY_EMAIL_VAR"
    pattern = "jd8-afe-ydv.{{ numeric(10) }}@synthetics.dtdg.co"
    example = "jd8-afe-ydv.4546132139@synthetics.dtdg.co"
  }

  variable {
    type = "global"
    name = "MY_GLOBAL_VAR"
    id   = "76636cd1-82e2-4aeb-9cfe-51366a8198a2"
  }
}
```

## Schema

### Required

- **locations** (Set of String, Required) Array of locations used to run the test. Refer to [Datadog documentation](https://docs.datadoghq.com/synthetics/api_test/#request) for available locations (e.g. `aws:eu-central-1`).
- **name** (String, Required) Name of Datadog synthetics test.
- **status** (String, Required) Define whether you want to start (`live`) or pause (`paused`) a Synthetic test. Allowed enum values: `live`, `paused`
- **type** (String, Required) Synthetics test type (`api` or `browser`).

### Optional

- **assertion** (Block List) Assertions used for the test. Multiple `assertion` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--assertion))
- **assertions** (List of Map of String, Optional, Deprecated) List of assertions.
- **browser_step** (Block List) Steps for browser tests. (see [below for nested schema](#nestedblock--browser_step))
- **browser_variable** (Block List) Variables used for a browser test steps. Multiple `variable` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--browser_variable))
- **config_variable** (Block List) Variables used for the test configuration. Multiple `config_variable` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--config_variable))
- **device_ids** (List of String, Optional) Array with the different device IDs used to run the test. Allowed enum values: `laptop_large`, `tablet`, `mobile_small` (only available for `browser` tests).
- **id** (String, Optional) The ID of this resource.
- **message** (String, Optional) A message to include with notifications for this synthetics test. Email notifications can be sent to specific users by using the same `@username` notation as events.
- **options** (Map of String, Optional, Deprecated)
- **options_list** (Block List, Max: 1) (see [below for nested schema](#nestedblock--options_list))
- **request** (Map of String, Optional, Deprecated) The synthetics test request. Required if `type = "api"`.
- **request_basicauth** (Block List, Max: 1) The HTTP basic authentication credentials. Exactly one nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--request_basicauth))
- **request_client_certificate** (Block List, Max: 1) Client certificate to use when performing the test request. Exactly one nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--request_client_certificate))
- **request_definition** (Block List, Max: 1) The synthetics test request. Required if `type = "api"`. (see [below for nested schema](#nestedblock--request_definition))
- **request_headers** (Map of String, Optional) Header name and value map.
- **request_query** (Map of String, Optional) Query arguments name and value map.
- **step** (Block List, Deprecated) Steps for browser tests. (see [below for nested schema](#nestedblock--step))
- **subtype** (String, Optional) When `type` is `api`, choose from `http`, `ssl`, `tcp` or `dns`. Defaults to `http`.
- **tags** (List of String, Optional) A list of tags to associate with your synthetics test. This can help you categorize and filter tests in the manage synthetics page of the UI. Default is an empty list (`[]`).
- **variable** (Block List, Deprecated) Variables used for a browser test steps. Multiple `browser_variable` blocks are allowed with the structure below. (see [below for nested schema](#nestedblock--variable))

### Read-only

- **monitor_id** (Number, Read-only) ID of the monitor associated with the Datadog synthetics test.

<a id="nestedblock--assertion"></a>
### Nested Schema for `assertion`

Required:

- **operator** (String, Required) Assertion operator. **Note** Only some combinations of `type` and `operator` are valid (please refer to [Datadog documentation](https://docs.datadoghq.com/synthetics/api_test/#validation)).
- **type** (String, Required) Type of assertion. Choose from `body`, `header`, `responseTime`, `statusCode`. **Note** Only some combinations of `type` and `operator` are valid (please refer to [Datadog documentation](https://docs.datadoghq.com/synthetics/api_test/#validation)).

Optional:

- **property** (String, Optional) If assertion type is `header`, this is the header name.
- **target** (String, Optional) Expected value. Depends on the assertion type, refer to [Datadog documentation](https://docs.datadoghq.com/synthetics/api_test/#validation) for details.
- **targetjsonpath** (Block List, Max: 1) Expected structure if `operator` is `validatesJSONPath`. Exactly one nested block is allowed with the structure below. (see [below for nested schema](#nestedblock--assertion--targetjsonpath))

<a id="nestedblock--assertion--targetjsonpath"></a>
### Nested Schema for `assertion.targetjsonpath`

Required:

- **jsonpath** (String, Required) The JSON path to assert.
- **operator** (String, Required) The specific operator to use on the path.
- **targetvalue** (String, Required) Expected matching value.



<a id="nestedblock--browser_step"></a>
### Nested Schema for `browser_step`

Required:

- **name** (String, Required) Name of the step.
- **params** (Block List, Min: 1, Max: 1) Parameters for the step. (see [below for nested schema](#nestedblock--browser_step--params))
- **type** (String, Required) Type of the step. Refer to [Datadog documentation](https://docs.datadoghq.com/api/v1/synthetics/#create-a-test) for the complete list of available types.

Optional:

- **allow_failure** (Boolean, Optional) Determines if the step should be allowed to fail.
- **force_element_update** (Boolean, Optional) Force update of the "element" parameter for the step
- **timeout** (Number, Optional) Used to override the default timeout of a step.

<a id="nestedblock--browser_step--params"></a>
### Nested Schema for `browser_step.params`

Optional:

- **attribute** (String, Optional) Name of the attribute to use for an "assert attribute" step.
- **check** (String, Optional) Check type to use for an assertion step.
- **click_type** (String, Optional) Type of click to use for a "click" step.
- **code** (String, Optional) Javascript code to use for the step.
- **delay** (Number, Optional) Delay between each key stroke for a "type test" step.
- **element** (String, Optional) Element to use for the step, json encoded string.
- **email** (String, Optional) Details of the email for an "assert email" step.
- **file** (String, Optional) For an "assert download" step.
- **files** (String, Optional) Details of the files for an "upload files" step, json encoded string.
- **modifiers** (List of String, Optional) Modifier to use for a "press key" step.
- **playing_tab_id** (String, Optional) ID of the tab to play the subtest.
- **request** (String, Optional) Request for an API step.
- **subtest_public_id** (String, Optional) ID of the Synthetics test to use as subtest.
- **value** (String, Optional) Value of the step.
- **variable** (Block List, Max: 1) Details of the variable to extract. (see [below for nested schema](#nestedblock--browser_step--params--variable))
- **with_click** (Boolean, Optional) For "file upload" steps.
- **x** (Number, Optional) X coordinates for a "scroll step".
- **y** (Number, Optional) Y coordinates for a "scroll step".

<a id="nestedblock--browser_step--params--variable"></a>
### Nested Schema for `browser_step.params.variable`

Optional:

- **example** (String, Optional)
- **name** (String, Optional) Name of the extracted variable.




<a id="nestedblock--browser_variable"></a>
### Nested Schema for `browser_variable`

Required:

- **name** (String, Required) Name of the variable.
- **type** (String, Required) Type of browser test variable. Allowed enum values: `element`, `email`, `global`, `javascript`, `text`.

Optional:

- **example** (String, Optional) Example for the variable.
- **id** (String, Optional) ID of the global variable to use. This is actually only used (and required) in the case of using a variable of type `global`.
- **pattern** (String, Optional) Pattern of the variable.


<a id="nestedblock--config_variable"></a>
### Nested Schema for `config_variable`

Required:

- **name** (String, Required) Name of the variable.
- **type** (String, Required) Type of test configuration variable. Allowed enum values: `text`.

Optional:

- **example** (String, Optional) Example for the variable.
- **pattern** (String, Optional) Pattern of the variable.


<a id="nestedblock--options_list"></a>
### Nested Schema for `options_list`

Optional:

- **accept_self_signed** (Boolean, Optional) For SSL test, whether or not the test should allow self signed certificates.
- **allow_insecure** (Boolean, Optional) Allows loading insecure content for an HTTP test.
- **follow_redirects** (Boolean, Optional) For API HTTP test, whether or not the test should follow redirects.
- **min_failure_duration** (Number, Optional) Minimum amount of time in failure required to trigger an alert. Default is `0`.
- **min_location_failed** (Number, Optional) Minimum number of locations in failure required to trigger an alert. Default is `1`.
- **monitor_options** (Block List, Max: 1) (see [below for nested schema](#nestedblock--options_list--monitor_options))
- **retry** (Block List, Max: 1) (see [below for nested schema](#nestedblock--options_list--retry))
- **tick_every** (Number, Optional) How often the test should run (in seconds). Current possible values are `900`, `1800`, `3600`, `21600`, `43200`, `86400`, `604800` plus `60` for API tests or `300` for browser tests.

<a id="nestedblock--options_list--monitor_options"></a>
### Nested Schema for `options_list.monitor_options`

Optional:

- **renotify_interval** (Number, Optional) Specify a renotification frequency.


<a id="nestedblock--options_list--retry"></a>
### Nested Schema for `options_list.retry`

Optional:

- **count** (Number, Optional) Number of retries needed to consider a location as failed before sending a notification alert.
- **interval** (Number, Optional) Interval between a failed test and the next retry in milliseconds.



<a id="nestedblock--request_basicauth"></a>
### Nested Schema for `request_basicauth`

Required:

- **password** (String, Required) Password for authentication.
- **username** (String, Required) Username for authentication.


<a id="nestedblock--request_client_certificate"></a>
### Nested Schema for `request_client_certificate`

Required:

- **cert** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--request_client_certificate--cert))
- **key** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--request_client_certificate--key))

<a id="nestedblock--request_client_certificate--cert"></a>
### Nested Schema for `request_client_certificate.cert`

Required:

- **content** (String, Required) Content of the certificate.

Optional:

- **filename** (String, Optional) File name for the certificate.


<a id="nestedblock--request_client_certificate--key"></a>
### Nested Schema for `request_client_certificate.key`

Required:

- **content** (String, Required) Content of the certificate.

Optional:

- **filename** (String, Optional) File name for the certificate.



<a id="nestedblock--request_definition"></a>
### Nested Schema for `request_definition`

Optional:

- **body** (String, Optional) The request body.
- **dns_server** (String, Optional) DNS server to use for DNS tests (`subtype = "dns"`).
- **host** (String, Optional) Host name to perform the test with.
- **method** (String, Optional) The HTTP method. One of `DELETE`, `GET`, `HEAD`, `OPTIONS`, `PATCH`, `POST`, `PUT`.
- **port** (Number, Optional) Port to use when performing the test.
- **timeout** (Number, Optional) Timeout in seconds for the test. Defaults to `60`.
- **url** (String, Optional) The URL to send the request to.


<a id="nestedblock--step"></a>
### Nested Schema for `step`

Required:

- **name** (String, Required) Name of the step.
- **params** (String, Required) Parameters for the step as JSON string.
- **type** (String, Required) Type of the step. Refer to [Datadog documentation](https://docs.datadoghq.com/api/v1/synthetics/#create-a-test) for the complete list of available types.

Optional:

- **allow_failure** (Boolean, Optional) Determines if the step should be allowed to fail.
- **force_element_update** (Boolean, Optional) Force update of the "element" parameter for the step
- **timeout** (Number, Optional) Used to override the default timeout of a step.


<a id="nestedblock--variable"></a>
### Nested Schema for `variable`

Required:

- **name** (String, Required) Name of the variable.
- **type** (String, Required) Type of browser test variable. Allowed enum values: `element`, `email`, `global`, `javascript`, `text`.

Optional:

- **example** (String, Optional) Example for the variable.
- **id** (String, Optional) ID of the global variable to use. This is actually only used (and required) in the case of using a variable of type `global`.
- **pattern** (String, Optional) Pattern of the variable.

## Import

Import is supported using the following syntax:

```shell
# Synthetics tests can be imported using their public string ID, e.g.
terraform import datadog_synthetics_test.fizz abc-123-xyz
```
