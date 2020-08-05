package datadog

import (
	"fmt"
	"github.com/jonboulle/clockwork"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogSyntheticsAPITest_importBasic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	testName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsAPITestConfig(testName),
			},
			{
				ResourceName:      "datadog_synthetics_test.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// Assertions will be imported into the new schema by default, but we can ignore them as users need to update the local config in this case
				ImportStateVerifyIgnore: []string{"assertions", "assertion"},
			},
		},
	})
}

func TestAccDatadogSyntheticsAPITest_importBasicNewAssertions(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	testName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsAPITestConfigNewAssertions(testName),
			},
			{
				ResourceName:      "datadog_synthetics_test.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsSSLTest_importBasic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	testName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsSSLTestConfig(testName),
			},
			{
				ResourceName:      "datadog_synthetics_test.ssl",
				ImportState:       true,
				ImportStateVerify: true,
				// Assertions will be imported into the new schema by default, but we can ignore them as users need to update the local config in this case
				ImportStateVerifyIgnore: []string{"assertions", "assertion"},
			},
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_importBasic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	testName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsBrowserTestConfig(testName),
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
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsAPITest_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStep(accProvider, clock, t),
			updateSyntheticsAPITestStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsAPITest_BasicNewAssertions(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStepNewAssertions(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsAPITest_UpdatedNewAssertions(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStepNewAssertions(accProvider, clock, t),
			updateSyntheticsAPITestStepNewAssertions(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsSSLTest_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsSSLTestStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsSSLTest_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsSSLTestStep(accProvider, clock, t),
			updateSyntheticsSSLTestStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStep(accProvider, clock, t),
			updateSyntheticsBrowserTestStep(accProvider, clock, t),
		},
	})
}

func createSyntheticsAPITestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsAPITestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "subtype", "http"),
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
				"datadog_synthetics_test.foo", "options.follow_redirects", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.retry_count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "name", testName),
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
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.foo", "monitor_id"),
		),
	}
}

func createSyntheticsAPITestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "foo" {
	type = "api"
	subtype = "http"

	request = {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		timeout = 30
	}
	request_headers = {
		Accept = "application/json"
		X-Datadog-Trace-ID = "1234566789"
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
	options = {
		tick_every = 60
		follow_redirects = true
		min_failure_duration = 0
		min_location_failed = 1
		retry_count = 1
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}`, uniq)
}

func createSyntheticsAPITestStepNewAssertions(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsAPITestConfigNewAssertions(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "subtype", "http"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_query.%", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_query.foo", "bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.0.username", "admin"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.0.password", "secret"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.type", "header"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.property", "content-type"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.operator", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.target", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.1.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.1.target", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.2.type", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.2.operator", "validatesJSONPath"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.2.targetjsonpath.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.2.targetjsonpath.0.jsonpath", "topKey"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.2.targetjsonpath.0.operator", "isNot"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.2.targetjsonpath.0.targetvalue", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.follow_redirects", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.retry_count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
		),
	}
}

func createSyntheticsAPITestConfigNewAssertions(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
	type = "api"
	subtype = "http"

	request = {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		timeout = 30
	}
	request_query = {
		foo = "bar"
	}
	request_basicauth {
		username = "admin"
		password = "secret"
	}
	request_headers = {
		Accept = "application/json"
		X-Datadog-Trace-ID = "1234566789"
	}

	assertion {
		type = "header"
		property = "content-type"
		operator = "contains"
		target = "application/json"
	}
	assertion {
		type = "statusCode"
		operator = "is"
		target = "200"
	}
	assertion {
		type = "body"
		operator = "validatesJSONPath"
		targetjsonpath {
			operator = "isNot"
			targetvalue = "0"
			jsonpath = "topKey"
		}
	}

	locations = [ "aws:eu-central-1" ]
	options = {
		tick_every = 60
		follow_redirects = true
		min_failure_duration = 0
		min_location_failed = 1
		retry_count = 1
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}`, uniq)
}

func updateSyntheticsAPITestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsAPITestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "subtype", "http"),
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
				"datadog_synthetics_test.foo", "options.follow_redirects", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.min_failure_duration", "10"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.retry_count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.retry_interval", "400"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "name", testName),
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
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.foo", "monitor_id"),
		),
	}
}

func updateSyntheticsAPITestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "foo" {
	type = "api"
	subtype = "http"

	request = {
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

	options = {
		tick_every = 900
		follow_redirects = false
		min_failure_duration = 10
		min_location_failed = 1
		retry_count = 1
		retry_interval = 400
	}

	name = "%s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"
}`, uniq)
}

func updateSyntheticsAPITestStepNewAssertions(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t) + "updated"
	return resource.TestStep{
		Config: updateSyntheticsAPITestConfigNewAssertions(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "subtype", "http"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request.url", "https://docs.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request.timeout", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.type", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.operator", "validatesJSONPath"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.targetjsonpath.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.targetjsonpath.0.jsonpath", "topKey"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.targetjsonpath.0.operator", "isNot"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.0.targetjsonpath.0.targetvalue", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.tick_every", "900"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.follow_redirects", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.min_failure_duration", "10"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.retry_count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "message", "Notify @pagerduty"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.1", "foo"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "status", "live"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
		),
	}
}

func updateSyntheticsAPITestConfigNewAssertions(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
	type = "api"
	subtype = "http"

	request = {
		method = "GET"
		url = "https://docs.datadoghq.com"
		timeout = 60
	}

	assertion {
		type = "body"
		operator = "validatesJSONPath"
		targetjsonpath {
			operator = "isNot"
			targetvalue = "0"
			jsonpath = "topKey"
		}
	}

	locations = [ "aws:eu-central-1" ]

	options = {
		tick_every = 900
		follow_redirects = false
		min_failure_duration = 10
		min_location_failed = 1
		retry_count = 1
	}

	name = "%s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"
}`, uniq)
}

func createSyntheticsSSLTestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsSSLTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "subtype", "ssl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request.host", "datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request.port", "443"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertions.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertions.0.type", "certificate"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertions.0.operator", "isInMoreThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertions.0.target", "30"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.accept_self_signed", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.ssl", "monitor_id"),
		),
	}
}

func createSyntheticsSSLTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "ssl" {
	type = "api"
	subtype = "ssl"

	request = {
		host = "datadoghq.com"
		port = 443
	}

	assertions = [
		{
			type = "certificate"
			operator = "isInMoreThan"
			target = 30
		}
	]

	locations = [ "aws:eu-central-1" ]
	options = {
		tick_every = 60
		accept_self_signed = true
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}`, uniq)
}

func updateSyntheticsSSLTestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsSSLTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "subtype", "ssl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request.host", "datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request.port", "443"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertions.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertions.0.type", "certificate"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertions.0.operator", "isInMoreThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertions.0.target", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.accept_self_signed", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "message", "Notify @pagerduty"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "tags.1", "foo"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "status", "live"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.ssl", "monitor_id"),
		),
	}
}

func updateSyntheticsSSLTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "ssl" {
	type = "api"
	subtype = "ssl"

	request = {
		host = "datadoghq.com"
		port = 443
	}

	assertions = [
		{
			type = "certificate"
			operator = "isInMoreThan"
			target = 60
		}
	]

	locations = [ "aws:eu-central-1" ]

	options = {
		tick_every = 60
		accept_self_signed = false
	}

	name = "%s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"
}`, uniq)
}

func createSyntheticsBrowserTestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsBrowserTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
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
				"datadog_synthetics_test.bar", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.1", "baz"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
		),
	}
}

func createSyntheticsBrowserTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
	type = "browser"

	request = {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		timeout = 30
	}
	request_headers = {
		Accept = "application/json"
		X-Datadog-Trace-ID = "123456789"
	}

	device_ids = [ "laptop_large", "mobile_small" ]
	locations = [ "aws:eu-central-1" ]
	options = {
		tick_every = 900
		min_failure_duration = 0
		min_location_failed = 1
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}`, uniq)
}

func updateSyntheticsBrowserTestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsBrowserTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
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
				"datadog_synthetics_test.bar", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "message", "Notify @pagerduty"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.1", "buz"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
		),
	}
}

func updateSyntheticsBrowserTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
	type = "browser"
	request = {
		method = "PUT"
		url = "https://docs.datadoghq.com"
		body = "this is an updated body"
		timeout = 60
	}
	request_headers = {
		Accept = "application/xml"
		X-Datadog-Trace-ID = "987654321"
	}
	device_ids = [ "laptop_large", "tablet" ]
	locations = [ "aws:eu-central-1" ]
	options = {
		tick_every = 1800
		min_failure_duration = 10
		min_location_failed = 1
	}
	name = "%s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "buz"]
	status = "live"
}`, uniq)
}

func testSyntheticsTestExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, _, err := datadogClientV1.SyntheticsApi.GetTest(authV1, r.Primary.ID).Execute(); err != nil {
				return fmt.Errorf("received an error retrieving synthetics test %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsTestIsDestroyed(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, _, err := datadogClientV1.SyntheticsApi.GetTest(authV1, r.Primary.ID).Execute(); err != nil {
				if strings.Contains(err.Error(), "404 Not Found") {
					continue
				}
				return fmt.Errorf("received an error retrieving synthetics test %s", err)
			}
			return fmt.Errorf("synthetics test still exists")
		}
		return nil
	}
}
