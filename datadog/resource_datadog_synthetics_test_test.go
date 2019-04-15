package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

func TestAccDatadogSyntheticsAPITest_importBasic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsAPITestConfig,
			},
			{
				ResourceName:      "datadog_synthetics_test.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_importBasic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsBrowserTestConfig,
			},
			{
				ResourceName:      "datadog_synthetics_test.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsAPITest_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed,
		Steps: []resource.TestStep{
			createSyntheticsAPITestStep,
		},
	})
}

func TestAccDatadogSyntheticsAPITest_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed,
		Steps: []resource.TestStep{
			createSyntheticsAPITestStep,
			updateSyntheticsAPITestStep,
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed,
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStep,
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed,
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStep,
			updateSyntheticsBrowserTestStep,
		},
	})
}

var createSyntheticsAPITestStep = resource.TestStep{
	Config: createSyntheticsAPITestConfig,
	Check: resource.ComposeTestCheckFunc(
		testSyntheticsTestExists(),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "type", "api"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "request.method", "GET"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "request.url", "https://www.datadoghq.com"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.#", "4"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.type", "header"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.property", "content-type"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.operator", "contains"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.target", "application/json"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.1.type", "statusCode"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.1.operator", "is"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.1.target", "200"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.2.type", "responseTime"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.2.operator", "lessThan"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.2.target", "2000"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.3.type", "body"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.3.operator", "doesNotContain"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.3.target", "terraform"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "locations.#", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "locations.0", "aws:eu-central-1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "options.tick_every", "60"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "options.min_failure_duration", "0"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "options.min_location_failed", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "name", "name for synthetics test foo"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "message", "Notify @datadog.user"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.#", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.0", "foo:bar"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.1", "baz"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "status", "paused"),
	),
}

const createSyntheticsAPITestConfig = `
resource "datadog_synthetics_test" "foo" {
	type = "api"

	request {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		timeout = 30
	}
	request_headers {
		"Accept" = "application/json"
		"X-Datadog-Trace-ID" = "1234566789"
	}

	assertions = [
		{
			type = "header"
			property = "content-type"
			operator = "contains"
			target = "application/json"
		},
		{
		  type = "statusCode"
		  operator = "is"
		  target = "200"
		},
		{
		  type = "responseTime"
		  operator = "lessThan"
			target = "2000"
		},
		{
		  type = "body"
		  operator = "doesNotContain"
		  target = "terraform"
		}
  ]

  locations = [ "aws:eu-central-1" ]
  options {
		tick_every = 60
		min_failure_duration = 0
		min_location_failed = 1
  }

  name = "name for synthetics test foo"
  message = "Notify @datadog.user"
  tags = ["foo:bar", "baz"]

  status = "paused"
}
`

var updateSyntheticsAPITestStep = resource.TestStep{
	Config: updateSyntheticsAPITestConfig,
	Check: resource.ComposeTestCheckFunc(
		testSyntheticsTestExists(),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "type", "api"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "request.method", "GET"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "request.url", "https://docs.datadoghq.com"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "request.timeout", "60"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.#", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.type", "statusCode"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.operator", "isNot"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "assertions.0.target", "500"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "locations.#", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "locations.0", "aws:eu-central-1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "options.tick_every", "900"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "options.min_failure_duration", "10"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "options.min_location_failed", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "name", "updated name"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "message", "Notify @pagerduty"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.#", "3"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.0", "foo:bar"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.1", "foo"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "tags.2", "env:test"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.foo", "status", "live"),
	),
}

const updateSyntheticsAPITestConfig = `
resource "datadog_synthetics_test" "foo" {
	type = "api"

	request {
	  method = "GET"
		url = "https://docs.datadoghq.com"
		timeout = 60
	}

	assertions = [
	  {
	    type = "statusCode"
	    operator = "isNot"
	    target = "500"
		}
	]

	locations = [ "aws:eu-central-1" ]

	options {
		tick_every = 900
		min_failure_duration = 10
		min_location_failed = 1
	}

	name = "updated name"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"
}
`

var createSyntheticsBrowserTestStep = resource.TestStep{
	Config: createSyntheticsBrowserTestConfig,
	Check: resource.ComposeTestCheckFunc(
		testSyntheticsTestExists(),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "type", "browser"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request.method", "GET"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request.url", "https://www.datadoghq.com"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request.body", "this is a body"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request.timeout", "30"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request_headers.%", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request_headers.Accept", "application/json"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request_headers.X-Datadog-Trace-ID", "123456789"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "device_ids.#", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "device_ids.0", "laptop_large"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "device_ids.1", "mobile_small"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "assertions.#", "0"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "locations.#", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "locations.0", "aws:eu-central-1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "options.tick_every", "900"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "options.min_failure_duration", "0"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "options.min_location_failed", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "name", "name for synthetics browser test bar"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "message", "Notify @datadog.user"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "tags.#", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "tags.0", "foo:bar"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "tags.1", "baz"),
	),
}

const createSyntheticsBrowserTestConfig = `
resource "datadog_synthetics_test" "bar" {
	type = "browser"

	request {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		timeout = 30
	}
	request_headers {
		"Accept" = "application/json"
		"X-Datadog-Trace-ID" = "123456789"
	}

	device_ids = [ "laptop_large", "mobile_small" ]
	locations = [ "aws:eu-central-1" ]
	options {
		tick_every = 900
		min_failure_duration = 0
		min_location_failed = 1
	}

	name = "name for synthetics browser test bar"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}
`

var updateSyntheticsBrowserTestStep = resource.TestStep{
	Config: updateSyntheticsBrowserTestConfig,
	Check: resource.ComposeTestCheckFunc(
		testSyntheticsTestExists(),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "type", "browser"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request.method", "PUT"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request.url", "https://docs.datadoghq.com"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request.body", "this is an updated body"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request.timeout", "60"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request_headers.%", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request_headers.Accept", "application/xml"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "request_headers.X-Datadog-Trace-ID", "987654321"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "device_ids.#", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "device_ids.0", "laptop_large"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "device_ids.1", "tablet"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "assertions.#", "0"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "locations.#", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "locations.0", "aws:eu-central-1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "options.tick_every", "1800"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "options.min_failure_duration", "10"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "options.min_location_failed", "1"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "name", "updated name for synthetics browser test bar"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "message", "Notify @pagerduty"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "tags.#", "2"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "tags.0", "foo:bar"),
		resource.TestCheckResourceAttr(
			"datadog_synthetics_test.bar", "tags.1", "buz"),
	),
}

const updateSyntheticsBrowserTestConfig = `
resource "datadog_synthetics_test" "bar" {
	type = "browser"
	request {
		method = "PUT"
		url = "https://docs.datadoghq.com"
		body = "this is an updated body"
		timeout = 60
	}
	request_headers {
		"Accept" = "application/xml"
		"X-Datadog-Trace-ID" = "987654321"
	}
	device_ids = [ "laptop_large", "tablet" ]
	locations = [ "aws:eu-central-1" ]
	options {
		tick_every = 1800
		min_failure_duration = 10
		min_location_failed = 1
	}
	name = "updated name for synthetics browser test bar"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "buz"]
	status = "live"
}
`

func testSyntheticsTestExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)

		for _, r := range s.RootModule().Resources {
			if _, err := client.GetSyntheticsTest(r.Primary.ID); err != nil {
				return fmt.Errorf("Received an error retrieving synthetics test %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsTestIsDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)

	for _, r := range s.RootModule().Resources {
		if _, err := client.GetSyntheticsTest(r.Primary.ID); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("Received an error retrieving synthetics test %s", err)
		}
		return fmt.Errorf("Synthetics test still exists")
	}
	return nil
}
