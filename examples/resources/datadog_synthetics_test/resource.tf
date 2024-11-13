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
    port = "443"
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
    port = "443"
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
      url    = "https://www.example.org"
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

  api_step {
    name    = "A gRPC health check on example.org"
    subtype = "grpc"

    assertion {
      type     = "statusCode"
      operator = "is"
      target   = "200"
    }

    request_definition {
      host      = "example.org"
      port      = "443"
      call_type = "healthcheck"
      service   = "greeter.Greeter"
    }
  }

  api_step {
    name    = "A gRPC behavior check on example.org"
    subtype = "grpc"

    assertion {
      type     = "statusCode"
      operator = "is"
      target   = "200"
    }

    request_definition {
      host      = "example.org"
      port      = "443"
      call_type = "unary"
      service   = "greeter.Greeter"
      method    = "SayHello"
      message   = "{\"name\": \"John\"}"

      plain_proto_file = <<EOT
syntax = "proto3";

package greeter;

// The greeting service definition.
service Greeter {
  // Sends a greeting
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}

// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}

// The response message containing the greetings
message HelloReply {
  string message = 1;
}
EOT      
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
    url    = "https://www.example.org"
  }

  browser_step {
    name = "Check current url"
    type = "assertCurrentUrl"
    params {
      check = "contains"
      value = "datadoghq"
    }
  }

  browser_step {
    name = "Test a downloaded file"
    type = "assertFileDownload"
    params {
      file = jsonencode(
        {
          md5 = "abcdef1234567890" // MD5 hash of the file
          sizeCheck = {
            type = "equals" // "equals", "greater", "greaterEquals", "lower", 
            // "lowerEquals", "notEquals", "between"
            value = 1
            // min   = 1      // only used for "between"
            // max   = 1      // only used for "between"
          }
          nameCheck = {
            type = "contains" // "contains", "equals", "isEmpty", "matchRegex", 
            // "notContains", "notIsEmpty", "notEquals", 
            // "notStartsWith", "startsWith"
            value = ".xls"
          }
        }
      )
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

# Example Usage (Synthetics Mobile test)
# Create a new Datadog Synthetics Mobile test starting on https://www.example.org
resource "datadog_synthetics_test" "test_mobile" {
	type = "mobile"
	config_variable {
			example = "123"
			name = "VARIABLE_NAME"
			pattern = "{{numeric(3)}}"
			type = "text"
			secure = false
		}
	config_initial_application_arguments = {
		test_process_argument = "test1"
	}
	device_ids = [ "synthetics:mobile:device:apple_iphone_14_plus_ios_16" ]
	locations = [ "aws:eu-central-1" ]
	mobile_options_list {
		min_failure_duration = 0
		retry {
			count = 0
			interval = 300
		}
		tick_every = 43200
		scheduling {
			timeframes {
        day = 5
        from = "07:00"
        to = "16:00"
      }
      timeframes  {
        day = 7
        from = "07:00"
        to = "16:00"
      }
			timezone = "UTC"
		}
		monitor_name = "%[1]s-monitor"
		monitor_options {
			renotify_interval = 10
			escalation_message = "test escalation message"
			renotify_occurrences = 3
			notification_preset_name = "show_all"
		}
		monitor_priority = 5
		restricted_roles = ["role1", "role2"]
		bindings {
      principal = [
        "org:8dee7c38-0000-aaaa-zzzz-8b5a08d3b091",
        "team:3a0cdd74-0000-aaaa-zzzz-da7ad0900002"
      ]
      relation = "editor"
    }
		ci {
			execution_rule = "blocking"
		}
		default_step_timeout = 10
		device_ids = ["synthetics:mobile:device:apple_iphone_14_plus_ios_16"]
		no_screenshot = true
		allow_application_crash = false
		disable_auto_accept_alert = true
		mobile_application {
			application_id = "5f055d15-0000-aaaa-zzzz-6739f83346aa"
      reference_id = "434d4719-0000-aaaa-zzzz-31082b544718"
      reference_type = "version"
		}
	}
	name = "%[1]s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]
	status = "paused"
	// monitor_id 
	mobile_step {
    name = "Tap on StaticText \"Tap\""
    params {
      element {
        context = "NATIVE_APP"
        view_name = "StaticText"
        context_type = "native"
        text_content = "Tap"
        multi_locator = {}
        relative_position {
          x = 0.07256155303030302
          y = 0.41522381756756754
        }
        user_locator {
          fail_test_on_cannot_locate = false
          values {
            type = "id"
            value = "some_id"
          }
        }
        element_description = "<XCUIElementTypeStaticText value=\"Tap\" name=\"Tap\" label=\"Tap\">"
      }
    }
    # "position": 0, // TODO this is not in the schema which seems like an oversight cuz this is the order of the steps
    timeout = 100
    type = "tap"
    public_id = "b9m-79b-idw" // TODO does need to be some unique id
    allow_failure = false
    is_critical = true
    no_screenshot = false
    # "exitIfSucceed": false // TODO this is not in the schema
    has_new_step_element = false
  }

  mobile_step {
    name = "Test View \"Tap\" content"
    params {
      check = "contains"
      value = "Tap"
      element {
        context = "NATIVE_APP"
        view_name = "View"
        context_type = "native"
        text_content = "Tap"
        multi_locator = {}
        relative_position {
          x = 0.27660448306074764
          y = 0.6841517857142857
        }
        user_locator {
          fail_test_on_cannot_locate = false
          values {
            type = "id"
            value = "some_id"
          }
        }
        element_description = "<XCUIElementTypeOther name=\"Tap\" label=\"Tap\">"
      }
    }
    # "position": 1, // TODO this is not in the schema which seems like an oversight cuz this is the order of the steps
    timeout = 100
    type = "assertElementContent"
    public_id = "uid-45h-9a6" // TODO does need to be some unique id
    allow_failure = false
    is_critical = true
    no_screenshot = false
    # exitIfSucceed = false  // TODO this is not in the schema
    has_new_step_element = false
  }
}


# Example Usage (GRPC API behavior check test)
# Create a new Datadog GRPC API test calling host example.org on port 443
# targeting service `greeter.Greeter` with the method `SayHello`
# and the message {"name": "John"}
resource "datadog_synthetics_test" "test_grpc_unary" {
  name      = "GRPC API behavior check test"
  type      = "api"
  subtype   = "grpc"
  status    = "live"
  locations = ["aws:eu-central-1"]
  tags      = ["foo:bar", "foo", "env:test"]

  request_definition {
    host      = "example.org"
    port      = "443"
    call_type = "unary"
    service   = "greeter.Greeter"
    method    = "SayHello"
    message   = "{\"name\": \"John\"}"

    plain_proto_file = <<EOT
syntax = "proto3";

package greeter;

// The greeting service definition.
service Greeter {
  // Sends a greeting
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}

// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}

// The response message containing the greetings
message HelloReply {
  string message = 1;
}
EOT
  }

  request_metadata = {
    header = "value"
  }

  assertion {
    type     = "responseTime"
    operator = "lessThan"
    target   = "2000"
  }

  assertion {
    operator = "is"
    type     = "grpcHealthcheckStatus"
    target   = 1
  }

  assertion {
    operator = "is"
    type     = "grpcProto"
    target   = "proto target"
  }

  assertion {
    operator = "is"
    property = "property"
    type     = "grpcMetadata"
    target   = "123"
  }

  options_list {
    tick_every = 900
  }
}

# Example Usage (GRPC API health check test)
# Create a new Datadog GRPC API test calling host example.org on port 443
# testing the overall health of the service
resource "datadog_synthetics_test" "test_grpc_health" {
  name      = "GRPC API health check test"
  type      = "api"
  subtype   = "grpc"
  status    = "live"
  locations = ["aws:eu-central-1"]
  tags      = ["foo:bar", "foo", "env:test"]

  request_definition {
    host      = "example.org"
    port      = "443"
    call_type = "healthcheck"
    service   = "greeter.Greeter"
  }

  assertion {
    type     = "responseTime"
    operator = "lessThan"
    target   = "2000"
  }

  assertion {
    operator = "is"
    type     = "grpcHealthcheckStatus"
    target   = 1
  }

  options_list {
    tick_every = 900
  }
}
