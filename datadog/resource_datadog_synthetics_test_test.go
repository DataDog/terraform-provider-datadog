package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jonboulle/clockwork"

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
				ImportStateVerifyIgnore: []string{"assertions", "assertion", "options", "options_list"},
			},
		},
	})
}

func TestAccDatadogSyntheticsAPITest_importBasicNewAssertionsOptions(t *testing.T) {
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
				Config: createSyntheticsAPITestConfigNewAssertionsOptions(testName),
			},
			{
				ResourceName:      "datadog_synthetics_test.bar",
				ImportState:       true,
				ImportStateVerify: true,
				// The request_client_certificate is not fully returned by the API so we can't verify it
				ImportStateVerifyIgnore: []string{"request_client_certificate"},
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
				ImportStateVerifyIgnore: []string{"assertions", "assertion", "options"},
			},
		},
	})
}

func TestAccDatadogSyntheticsTCPTest_importBasic(t *testing.T) {
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
				Config: createSyntheticsTCPTestConfig(testName),
			},
			{
				ResourceName:      "datadog_synthetics_test.tcp",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsDNSTest_importBasic(t *testing.T) {
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
				Config: createSyntheticsDNSTestConfig(testName),
			},
			{
				ResourceName:      "datadog_synthetics_test.dns",
				ImportState:       true,
				ImportStateVerify: true,
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
				ResourceName:            "datadog_synthetics_test.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"options_list", "browser_variable"},
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

func TestAccDatadogSyntheticsAPITest_BasicNewAssertionsOptions(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStepNewAssertionsOptions(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsAPITest_UpdatedNewAssertionsOptions(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStepNewAssertionsOptions(accProvider, clock, t),
			updateSyntheticsAPITestStepNewAssertionsOptions(accProvider, clock, t),
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

func TestAccDatadogSyntheticsSSLMissingTagsAttributeTest_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsSSLMissingTagsAttributeTestStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsTCPTest_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsTCPTestStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsTCPTest_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsTCPTestStep(accProvider, clock, t),
			updateSyntheticsTCPTestStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsDNSTest_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsDNSTestStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsDNSTest_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsDNSTestStep(accProvider, clock, t),
			updateSyntheticsDNSTestStep(accProvider, clock, t),
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

func TestAccDatadogSyntheticsBrowserTestBrowserVariables_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestBrowserVariablesStep(accProvider, clock, t),
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
				"datadog_synthetics_test.foo", "options.allow_insecure", "true"),
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
				"datadog_synthetics_test.foo", "options_list.#", "0"),
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
		allow_insecure = true
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

func createSyntheticsAPITestStepNewAssertionsOptions(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsAPITestConfigNewAssertionsOptions(testName),
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
				"datadog_synthetics_test.bar", "request_client_certificate.0.cert.0.content", "content-certificate"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_client_certificate.0.cert.0.filename", "Provided in Terraform config"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_client_certificate.0.key.0.content", "content-key"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_client_certificate.0.key.0.filename", "key"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.#", "7"),
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
				"datadog_synthetics_test.bar", "assertion.3.type", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.3.operator", "validatesJSONPath"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.3.targetjsonpath.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.3.targetjsonpath.0.jsonpath", "something"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.3.targetjsonpath.0.operator", "moreThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.3.targetjsonpath.0.targetvalue", "5"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.4.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.4.operator", "isNot"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.4.target", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.5.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.5.operator", "matches"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.5.target", "20[04]"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.6.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.6.operator", "doesNotMatch"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.6.target", "20[04]"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.follow_redirects", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.monitor_options.0.renotify_interval", "100"),
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

func createSyntheticsAPITestConfigNewAssertionsOptions(uniq string) string {
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
	request_client_certificate {
		cert {
			content = "content-certificate"
		}
		key {
			content = "content-key"
			filename = "key"
		}
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
	assertion {
		type = "body"
		operator = "validatesJSONPath"
		targetjsonpath {
			operator = "moreThan"
			targetvalue = "5"
			jsonpath = "something"
		}
	}
	assertion {
		type = "statusCode"
		operator = "isNot"
		target = "200"
	}
	assertion {
		type = "statusCode"
		operator = "matches"
		target = "20[04]"
	}
	assertion {
		type = "statusCode"
		operator = "doesNotMatch"
		target = "20[04]"
	}

	locations = [ "aws:eu-central-1" ]
	options_list {
		tick_every = 60
		follow_redirects = true
		min_failure_duration = 0
		min_location_failed = 1

		monitor_options {
			renotify_interval = 100
		}
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
				"datadog_synthetics_test.foo", "options_list.0.tick_every", "900"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.follow_redirects", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.min_failure_duration", "10"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.retry.0.count", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.retry.0.interval", "500"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.monitor_options.0.renotify_interval", "100"),
			// Make sure the legacy attribute isn't set anymore
			resource.TestCheckNoResourceAttr("datadog_synthetics_test.foo", "options.tick_every"),
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

	options_list {
		tick_every = 900
		follow_redirects = false
		min_failure_duration = 10
		min_location_failed = 1

		retry {
			count = 3
			interval = 500
		}

		monitor_options {
			renotify_interval = 100
		}
	}

	name = "%s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"
}`, uniq)
}

func updateSyntheticsAPITestStepNewAssertionsOptions(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t) + "updated"
	return resource.TestStep{
		Config: updateSyntheticsAPITestConfigNewAssertionsOptions(testName),
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
				"datadog_synthetics_test.bar", "options_list.0.tick_every", "900"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.follow_redirects", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_failure_duration", "10"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.monitor_options.0.renotify_interval", "120"),
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

func updateSyntheticsAPITestConfigNewAssertionsOptions(uniq string) string {
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

	options_list {
		tick_every = 900
		follow_redirects = false
		min_failure_duration = 10
		min_location_failed = 1

		monitor_options {
			renotify_interval = 120
		}
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
				"datadog_synthetics_test.ssl", "tags.#", "0"),
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
	tags = []

	status = "paused"
}`, uniq)
}

func createSyntheticsSSLMissingTagsAttributeTestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsSSLMissingTagsAttributeTestConfig(testName),
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
				"datadog_synthetics_test.ssl", "tags.#", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.ssl", "monitor_id"),
		),
	}
}

func createSyntheticsSSLMissingTagsAttributeTestConfig(uniq string) string {
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

func createSyntheticsTCPTestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsTCPTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "subtype", "tcp"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "request.host", "agent-intake.logs.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "request.port", "443"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.0.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.0.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.0.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "options_list.0.tick_every", "60"),
			// Make sure the legacy attribute isn't set anymore
			resource.TestCheckNoResourceAttr("datadog_synthetics_test.tcp", "options.tick_every"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.tcp", "monitor_id"),
		),
	}
}

func createSyntheticsTCPTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "tcp" {
	type = "api"
	subtype = "tcp"

	request = {
		host = "agent-intake.logs.datadoghq.com"
		port = 443
	}

	assertion {
		type = "responseTime"
		operator = "lessThan"
		target = 2000
	}

	locations = [ "aws:eu-central-1" ]
	options_list {
		tick_every = 60
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}`, uniq)
}

func updateSyntheticsTCPTestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsTCPTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "subtype", "tcp"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "request.host", "agent-intake.logs.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "request.port", "443"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.0.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.0.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.0.target", "3000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "options_list.0.tick_every", "300"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "status", "live"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.tcp", "monitor_id"),
		),
	}
}

func updateSyntheticsTCPTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "tcp" {
	type = "api"
	subtype = "tcp"

	request = {
		host = "agent-intake.logs.datadoghq.com"
		port = 443
	}

	assertion {
		type = "responseTime"
		operator = "lessThan"
		target = 3000
	  }

	locations = [ "aws:eu-central-1" ]
	options_list {
		tick_every = 300
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz", "env:test"]

	status = "live"
}`, uniq)
}

func createSyntheticsDNSTestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsDNSTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "subtype", "dns"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "request.host", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "request.dns_server", "8.8.8.8"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.0.type", "recordSome"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.0.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.0.property", "A"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.0.target", "0.0.0.0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.dns", "monitor_id"),
		),
	}
}

func createSyntheticsDNSTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "dns" {
	type = "api"
	subtype = "dns"

	request = {
		host = "https://www.datadoghq.com"
		dns_server = "8.8.8.8"
	}

	assertion {
		type = "recordSome"
		operator = "is"
		property = "A"
		target = "0.0.0.0"
	}

	locations = [ "aws:eu-central-1" ]
	options_list {
		tick_every = 60
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}`, uniq)
}

func updateSyntheticsDNSTestStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsDNSTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "subtype", "dns"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "request.host", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "request.dns_server", "8.8.8.8"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.0.type", "recordEvery"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.0.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.0.property", "A"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "assertion.0.target", "1.1.1.1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "options_list.0.tick_every", "300"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "status", "live"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.dns", "monitor_id"),
		),
	}
}

func updateSyntheticsDNSTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "dns" {
	type = "api"
	subtype = "dns"

	request = {
		host = "https://www.datadoghq.com"
		dns_server = "8.8.8.8"
	}

	assertion {
		type = "recordEvery"
		operator = "is"
		property = "A"
		target = "1.1.1.1"
	  }

	locations = [ "aws:eu-central-1" ]
	options_list {
		tick_every = 300
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz", "env:test"]

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
				"datadog_synthetics_test.bar", "options_list.0.tick_every", "900"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.retry.0.count", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.retry.0.interval", "300"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.monitor_options.0.renotify_interval", "100"),
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
				"datadog_synthetics_test.bar", "step.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.0.name", "first step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.0.type", "assertCurrentUrl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.0.params", "{\"check\":\"contains\",\"value\":\"content\"}"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "variable.0.type", "text"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "variable.0.name", "MY_PATTERN_VAR"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "variable.0.pattern", "{{numeric(3)}}"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "variable.0.example", "597"),
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
	options_list {
		tick_every = 900
		min_failure_duration = 0
		min_location_failed = 1

		retry {
			count = 2
			interval = 300
		}

		monitor_options {
			renotify_interval = 100
		}
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"

	step {
	    name = "first step"
	    type = "assertCurrentUrl"
	    params = jsonencode({
	        "check": "contains",
	        "value": "content"
	    })
	}

	variable {
		type = "text"
		name = "MY_PATTERN_VAR"
		pattern = "{{numeric(3)}}"
		example = "597"
	}
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
				"datadog_synthetics_test.bar", "options_list.0.tick_every", "1800"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_failure_duration", "10"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.retry.0.count", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.retry.0.interval", "500"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.monitor_options.0.renotify_interval", "120"),
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
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.0.name", "first step updated"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.0.type", "assertCurrentUrl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.0.params", "{\"check\":\"contains\",\"value\":\"content\"}"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.1.name", "press key step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.1.type", "pressKey"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.1.params", "{\"value\":\"1\"}"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "variable.0.type", "text"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "variable.0.name", "MY_PATTERN_VAR"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "variable.0.pattern", "{{numeric(4)}}"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "variable.0.example", "5970"),
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
	options_list {
		tick_every = 1800
		min_failure_duration = 10
		min_location_failed = 1

		retry {
			count = 3
			interval = 500
		}

		monitor_options {
			renotify_interval = 120
		}
	}
	name = "%s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "buz"]
	status = "live"

	step {
	    name = "first step updated"
	    type = "assertCurrentUrl"

	    params = jsonencode({
	        "check": "contains",
	        "value": "content"
	    })
	}

	step {
	    name = "press key step"
	    type = "pressKey"

	    params = jsonencode({
	        "value": "1"
	    })
	}

	variable {
		type = "text"
		name = "MY_PATTERN_VAR"
		pattern = "{{numeric(4)}}"
		example = "5970"
	}
}`, uniq)
}

func createSyntheticsBrowserTestBrowserVariablesStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	testName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsBrowserTestBrowserVariablesConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "browser"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "device_ids.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "device_ids.0", "laptop_large"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertions.#", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.tick_every", "900"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.retry.0.count", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.retry.0.interval", "300"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.0.name", "first step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.0.type", "assertCurrentUrl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "step.0.params", "{\"check\":\"contains\",\"value\":\"content\"}"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.0.type", "text"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.0.name", "MY_PATTERN_VAR"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.0.pattern", "{{numeric(3)}}"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.0.example", "597"),
		),
	}
}

func createSyntheticsBrowserTestBrowserVariablesConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
       type = "browser"

       request = {
               method = "GET"
               url = "https://www.datadoghq.com"
       }

       device_ids = [ "laptop_large" ]
       locations = [ "aws:eu-central-1" ]
       options_list {
               tick_every = 900
               min_failure_duration = 0
               min_location_failed = 1

               retry {
                       count = 2
                       interval = 300
               }
       }

       name = "%s"
       message = "Notify @datadog.user"
       tags = ["foo:bar"]

       status = "paused"

       step {
           name = "first step"
           type = "assertCurrentUrl"
           params = jsonencode({
               "check": "contains",
               "value": "content"
           })
       }

       browser_variable {
               type = "text"
               name = "MY_PATTERN_VAR"
               pattern = "{{numeric(3)}}"
               example = "597"
       }
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
