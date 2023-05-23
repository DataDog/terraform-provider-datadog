# Example Usage (Synthetics API test)
# Create a new Datadog Synthetics API/HTTP test on https://www.example.org
resource "datadog_synthetics_test" "test_uptime" {
  name      = "An Uptime test on example.org"
  type      = "api"
  subtype   = "http"
  status    = "live"
  message   = "Notify @pagerduty"
  locations = ["aws:eu-central-1"]
  tags      = ["foo:bar", "foo", "env:test"]

  request_definition {
    method = "GET"
    url    = "https://www.example.org"
  }

  request_headers = {
    Content-Type = "application/json"
  }

  assertion {
    type     = "statusCode"
    operator = "is"
    target   = "200"
  }

  options_list {
    tick_every = 900
    retry {
      count    = 2
      interval = 300
    }
    monitor_options {
      renotify_interval = 120
    }
  }
}


# Example Usage (Authenticated API test)
# Create a new Datadog Synthetics API/HTTP test on https://www.example.org
resource "datadog_synthetics_test" "test_api" {
  name      = "An API test on example.org"
  type      = "api"
  subtype   = "http"
  status    = "live"
  message   = "Notify @pagerduty"
  locations = ["aws:eu-central-1"]
  tags      = ["foo:bar", "foo", "env:test"]

  request_definition {
    method = "GET"
    url    = "https://www.example.org"
  }

  request_headers = {
    Content-Type   = "application/json"
    Authentication = "Token: 1234566789"
  }

  assertion {
    type     = "statusCode"
    operator = "is"
    target   = "200"
  }

  options_list {
    tick_every = 900
    retry {
      count    = 2
      interval = 300
    }
    monitor_options {
      renotify_interval = 120
    }
  }
}


# Example Usage (Synthetics SSL test)
# Create a new Datadog Synthetics API/SSL test on example.org
resource "datadog_synthetics_test" "test_ssl" {
  name      = "An API test on example.org"
  type      = "api"
  subtype   = "ssl"
  status    = "live"
  message   = "Notify @pagerduty"
  locations = ["aws:eu-central-1"]
  tags      = ["foo:bar", "foo", "env:test"]

  request_definition {
    host = "example.org"
    port = 443
  }

  assertion {
    type     = "certificate"
    operator = "isInMoreThan"
    target   = 30
  }

  options_list {
    tick_every         = 900
    accept_self_signed = true
  }
}


# Example Usage (Synthetics TCP test)
# Create a new Datadog Synthetics API/TCP test on example.org
resource "datadog_synthetics_test" "test_tcp" {
  name      = "An API test on example.org"
  type      = "api"
  subtype   = "tcp"
  status    = "live"
  message   = "Notify @pagerduty"
  locations = ["aws:eu-central-1"]
  tags      = ["foo:bar", "foo", "env:test"]

  request_definition {
    host = "example.org"
    port = 443
  }

  assertion {
    type     = "responseTime"
    operator = "lessThan"
    target   = 2000
  }

  config_variable {
    type = "global"
    name = "MY_GLOBAL_VAR"
    id   = "76636cd1-82e2-4aeb-9cfe-51366a8198a2"
  }

  options_list {
    tick_every = 900
  }
}


# Example Usage (Synthetics DNS test)
# Create a new Datadog Synthetics API/DNS test on example.org
resource "datadog_synthetics_test" "test_dns" {
  name      = "An API test on example.org"
  type      = "api"
  subtype   = "dns"
  status    = "live"
  message   = "Notify @pagerduty"
  locations = ["aws:eu-central-1"]
  tags      = ["foo:bar", "foo", "env:test"]

  request_definition {
    host = "example.org"
  }

  assertion {
    type     = "recordSome"
    operator = "is"
    property = "A"
    target   = "0.0.0.0"
  }

  options_list {
    tick_every = 900
  }
}


# Example Usage (Synthetics Multistep API test)
# Create a new Datadog Synthetics Multistep API test
resource "datadog_synthetics_test" "test_multi_step" {
  name      = "Multistep API test"
  type      = "api"
  subtype   = "multi"
  status    = "live"
  locations = ["aws:eu-central-1"]
  tags      = ["foo:bar", "foo", "env:test"]

  api_step {
    name    = "An API test on example.org"
    subtype = "http"

    assertion {
      type     = "statusCode"
      operator = "is"
      target   = "200"
    }

    request_definition {
      method = "GET"
      url    = "https://example.org"
    }

    request_headers = {
      Content-Type   = "application/json"
      Authentication = "Token: 1234566789"
    }
  }

  api_step {
    name    = "An API test on example.org"
    subtype = "http"

    assertion {
      type     = "statusCode"
      operator = "is"
      target   = "200"
    }

    request_definition {
      method = "GET"
      url    = "http://example.org"
    }
  }

  options_list {
    tick_every         = 900
    accept_self_signed = true
  }
}


# Example Usage (Synthetics Browser test)
# Create a new Datadog Synthetics Browser test starting on https://www.example.org
resource "datadog_synthetics_test" "test_browser" {
  name       = "A Browser test on example.org"
  type       = "browser"
  status     = "paused"
  message    = "Notify @qa"
  device_ids = ["laptop_large"]
  locations  = ["aws:eu-central-1"]
  tags       = []

  request_definition {
    method = "GET"
    url    = "https://app.datadoghq.com"
  }

  browser_step {
    name = "Check current url"
    type = "assertCurrentUrl"
    params {
      check = "contains"
      value = "datadoghq"
    }
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

  options_list {
    tick_every = 3600
  }
}
