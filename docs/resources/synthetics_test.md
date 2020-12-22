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

- **locations** (List of String, Required)
- **name** (String, Required)
- **request** (Map of String, Required)
- **status** (String, Required)
- **type** (String, Required)

### Optional

- **assertion** (Block List) (see [below for nested schema](#nestedblock--assertion))
- **assertions** (List of Map of String, Optional, Deprecated)
- **device_ids** (List of String, Optional)
- **id** (String, Optional) The ID of this resource.
- **message** (String, Optional)
- **options** (Map of String, Optional, Deprecated)
- **options_list** (Block List, Max: 1) (see [below for nested schema](#nestedblock--options_list))
- **request_basicauth** (Block List, Max: 1) (see [below for nested schema](#nestedblock--request_basicauth))
- **request_client_certificate** (Block List, Max: 1) (see [below for nested schema](#nestedblock--request_client_certificate))
- **request_headers** (Map of String, Optional)
- **request_query** (Map of String, Optional)
- **step** (Block List) (see [below for nested schema](#nestedblock--step))
- **subtype** (String, Optional)
- **tags** (List of String, Optional)
- **variable** (Block List) (see [below for nested schema](#nestedblock--variable))

### Read-only

- **monitor_id** (Number, Read-only)

<a id="nestedblock--assertion"></a>
### Nested Schema for `assertion`

Required:

- **operator** (String, Required)
- **type** (String, Required)

Optional:

- **property** (String, Optional)
- **target** (String, Optional)
- **targetjsonpath** (Block List, Max: 1) (see [below for nested schema](#nestedblock--assertion--targetjsonpath))

<a id="nestedblock--assertion--targetjsonpath"></a>
### Nested Schema for `assertion.targetjsonpath`

Required:

- **jsonpath** (String, Required)
- **operator** (String, Required)
- **targetvalue** (String, Required)



<a id="nestedblock--options_list"></a>
### Nested Schema for `options_list`

Optional:

- **accept_self_signed** (Boolean, Optional)
- **allow_insecure** (Boolean, Optional)
- **follow_redirects** (Boolean, Optional)
- **min_failure_duration** (Number, Optional)
- **min_location_failed** (Number, Optional)
- **monitor_options** (Block List, Max: 1) (see [below for nested schema](#nestedblock--options_list--monitor_options))
- **retry** (Block List, Max: 1) (see [below for nested schema](#nestedblock--options_list--retry))
- **tick_every** (Number, Optional)

<a id="nestedblock--options_list--monitor_options"></a>
### Nested Schema for `options_list.monitor_options`

Optional:

- **renotify_interval** (Number, Optional)


<a id="nestedblock--options_list--retry"></a>
### Nested Schema for `options_list.retry`

Optional:

- **count** (Number, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--request_basicauth"></a>
### Nested Schema for `request_basicauth`

Required:

- **password** (String, Required)
- **username** (String, Required)


<a id="nestedblock--request_client_certificate"></a>
### Nested Schema for `request_client_certificate`

Required:

- **cert** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--request_client_certificate--cert))
- **key** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--request_client_certificate--key))

<a id="nestedblock--request_client_certificate--cert"></a>
### Nested Schema for `request_client_certificate.cert`

Required:

- **content** (String, Required)

Optional:

- **filename** (String, Optional)


<a id="nestedblock--request_client_certificate--key"></a>
### Nested Schema for `request_client_certificate.key`

Required:

- **content** (String, Required)

Optional:

- **filename** (String, Optional)



<a id="nestedblock--step"></a>
### Nested Schema for `step`

Required:

- **name** (String, Required)
- **params** (String, Required)
- **type** (String, Required)

Optional:

- **allow_failure** (Boolean, Optional)
- **timeout** (Number, Optional)


<a id="nestedblock--variable"></a>
### Nested Schema for `variable`

Required:

- **name** (String, Required)
- **type** (String, Required)

Optional:

- **example** (String, Optional)
- **id** (String, Optional) The ID of this resource.
- **pattern** (String, Optional)

## Import

Import is supported using the following syntax:

```shell
# Synthetics tests can be imported using their public string ID, e.g.
terraform import datadog_synthetics_test.fizz abc-123-xyz
```
