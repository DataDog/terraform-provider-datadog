package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

// TODO: add options.retry and options.monitor_options tests once options become a TypeList on next Major version

func testAccDatadogSyntheticsAPITestStateUpgradeV0(name string, accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Resource not found %s", name)
		}
		if resource.Primary.ID == "" {
			return fmt.Errorf("Resource has no ID")
		}
		rawState := map[string]interface{}{
			"options.follow_redirects":     resource.Primary.Attributes["options.follow_redirects"],
			"options.min_failure_duration": resource.Primary.Attributes["options.min_failure_duration"],
			"options.min_location_failed":  resource.Primary.Attributes["options.min_location_failed"],
			"options.tick_every":           resource.Primary.Attributes["options.tick_every"],
			"options.accept_self_signed":   resource.Primary.Attributes["options.accept_self_signed"],
		}
		updatedRawState, err := resourceDatadogSyntheticsTestStateUpgradeV0(rawState, accProvider.Meta())
		optionsKV := map[string]string{
			"options.follow_redirects":     "options.0.follow_redirects",
			"options.min_failure_duration": "options.0.min_failure_duration",
			"options.min_location_failed":  "options.0.min_location_failed",
			"options.tick_every":           "options.0.tick_every",
			"options.accept_self_signed":   "options.0.accept_self_signed",
		}
		if err != nil {
			return err
		}
		for v0Field, v1Field := range optionsKV {
			if rawState[v0Field] != updatedRawState[v1Field] {
				s := fmt.Errorf("Error: fields are not equal after upgrade %s:%s -> %s:%s", v0Field, rawState[v0Field], v1Field, updatedRawState[v1Field])
				fmt.Println(s)
			}
		}
		return nil
	}
}

func TestAccDatadogSyntheticsAPITest_importBasic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
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

func TestAccDatadogSyntheticsAPITestV0_importBasic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsAPITestConfigV0,
			},
			{
				ResourceName:      "datadog_synthetics_test_V0.pepito",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsSSLTest_importBasic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsSSLTestConfig,
			},
			{
				ResourceName:      "datadog_synthetics_test.ssl",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_importBasic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
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
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStep(accProvider),
		},
	})
}

func TestAccDatadogSyntheticsAPITest_StateUpgradeV0(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStepV0(accProvider),
			updateSyntheticsAPITestStepV0(accProvider),
		},
	})
}

func TestAccDatadogSyntheticsAPITest_Updated(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStep(accProvider),
			updateSyntheticsAPITestStep(accProvider),
		},
	})
}

func TestAccDatadogSyntheticsSSLTest_Basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsSSLTestStep(accProvider),
		},
	})
}

func TestAccDatadogSyntheticsSSLTest_Updated(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsSSLTestStep(accProvider),
			updateSyntheticsSSLTestStep(accProvider),
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_Basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStep(accProvider),
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_Updated(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStep(accProvider),
			updateSyntheticsBrowserTestStep(accProvider),
		},
	})
}

func createSyntheticsAPITestStepV0(accProvider *schema.Provider) resource.TestStep {
	return resource.TestStep{
		Config: createSyntheticsAPITestConfigV0,
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "subtype", "http"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "request.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "request.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.#", "4"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.0.type", "header"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.0.property", "content-type"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.0.operator", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.0.target", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.1.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.1.target", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.2.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.2.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.2.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.3.type", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.3.operator", "doesNotContain"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.3.target", "terraform"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "options.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "options.follow_redirects", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "options.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "options.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "name", "name for synthetics test foo"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test_V0.pepito", "monitor_id"),
		),
	}
}

const createSyntheticsAPITestConfigV0 = `
resource "datadog_synthetics_test_V0" "pepito" {
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
        }

        name = "name for synthetics test foo"
        message = "Notify @datadog.user"
        tags = ["foo:bar", "baz"]

        status = "paused"
}
`

func updateSyntheticsAPITestStepV0(accProvider *schema.Provider) resource.TestStep {
	return resource.TestStep{
		Config: updateSyntheticsAPITestConfigV0,
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "subtype", "http"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "request.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "request.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.#", "4"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.0.type", "header"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.0.property", "content-type"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.0.operator", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.0.target", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.1.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.1.target", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.2.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.2.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.2.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.3.type", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.3.operator", "doesNotContain"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "assertions.3.target", "terraform"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "options.tick_every", "300"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "options.follow_redirects", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "options.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "options.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "name", "name for synthetics test foo"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test_V0.pepito", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test_V0.pepito", "monitor_id"),
			testAccDatadogSyntheticsAPITestStateUpgradeV0("datadog_synthetics_test_V0.pepito", accProvider),
		),
	}
}

const updateSyntheticsAPITestConfigV0 = `
resource "datadog_synthetics_test_V0" "pepito" {
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
                tick_every = 300
                follow_redirects = true
                min_failure_duration = 0
                min_location_failed = 1
        }

        name = "name for synthetics test foo"
        message = "Notify @datadog.user"
        tags = ["foo:bar", "baz"]

        status = "paused"
}
`

func createSyntheticsAPITestStep(accProvider *schema.Provider) resource.TestStep {
	return resource.TestStep{
		Config: createSyntheticsAPITestConfig,
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
				"datadog_synthetics_test.foo", "options.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.follow_redirects", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.retry.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.retry.count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.retry.interval", "100"),
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
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.foo", "monitor_id"),
		),
	}
}

const createSyntheticsAPITestConfig = `
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
	options {
		tick_every = 60
		follow_redirects = true
		min_failure_duration = 0
		min_location_failed = 1

                retry = {
                    count = 1
                    interval = 100
                }

	}

	name = "name for synthetics test foo"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}
`

func updateSyntheticsAPITestStep(accProvider *schema.Provider) resource.TestStep {
	return resource.TestStep{
		Config: updateSyntheticsAPITestConfig,
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
				"datadog_synthetics_test.foo", "options.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.tick_every", "900"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.follow_redirects", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.min_failure_duration", "10"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.monitor_options.%", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options.0.monitor_options.renotify_interval", "88"),
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
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.foo", "monitor_id"),
		),
	}
}

const updateSyntheticsAPITestConfig = `
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

	options {
		tick_every = 900
		follow_redirects = false
		min_failure_duration = 10
		min_location_failed = 1

                monitor_options = {
                    renotify_interval = 88
                }
	}

	name = "updated name"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"
}
`

func createSyntheticsSSLTestStep(accProvider *schema.Provider) resource.TestStep {
	return resource.TestStep{
		Config: createSyntheticsSSLTestConfig,
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
				"datadog_synthetics_test.ssl", "options.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.accept_self_signed", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.retry.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.retry.count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.retry.interval", "88"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.monitor_options.%", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.monitor_options.renotify_interval", "100"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "name", "name for synthetics test ssl"),
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

const createSyntheticsSSLTestConfig = `
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
	options {
		tick_every = 60
		accept_self_signed = true

                monitor_options = {
                    renotify_interval = 100
                }
                retry = {
                    count = 1
                    interval = 88
                }
	}

	name = "name for synthetics test ssl"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}
`

func updateSyntheticsSSLTestStep(accProvider *schema.Provider) resource.TestStep {
	return resource.TestStep{
		Config: updateSyntheticsSSLTestConfig,
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
				"datadog_synthetics_test.ssl", "options.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.accept_self_signed", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.retry.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.retry.count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.retry.interval", "70"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.monitor_options.%", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options.0.monitor_options.renotify_interval", "800"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "name", "updated name"),
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

const updateSyntheticsSSLTestConfig = `
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

	options {
		tick_every = 60
		accept_self_signed = false

               monitor_options = {
                    renotify_interval = 800
                }
                retry = {
                    count = 1
                    interval = 70
                }
	}

	name = "updated name"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"
}
`

func createSyntheticsBrowserTestStep(accProvider *schema.Provider) resource.TestStep {
	return resource.TestStep{
		Config: createSyntheticsBrowserTestConfig,
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
				"datadog_synthetics_test.bar", "options.0.tick_every", "900"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.monitor_options.%", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.monitor_options.renotify_interval", "10"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.retry.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.retry.count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.retry.interval", "100"),
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
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
		),
	}
}

const createSyntheticsBrowserTestConfig = `
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
	options {
		tick_every = 900
		min_failure_duration = 0
		min_location_failed = 1

                monitor_options = {
                    renotify_interval = 10
                }
                retry = {
                    count = 1
                    interval = 100
                }
	}

	name = "name for synthetics browser test bar"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}
`

func updateSyntheticsBrowserTestStep(accProvider *schema.Provider) resource.TestStep {
	return resource.TestStep{
		Config: updateSyntheticsBrowserTestConfig,
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
				"datadog_synthetics_test.bar", "options.0.tick_every", "1800"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.min_failure_duration", "10"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.monitor_options.%", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.monitor_options.renotify_interval", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.retry.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.retry.count", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options.0.retry.interval", "20"),
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
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
		),
	}
}

const updateSyntheticsBrowserTestConfig = `
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
	options {
		tick_every = 1800
		min_failure_duration = 10
		min_location_failed = 1

                monitor_options = {
                    renotify_interval = 2
                }
                retry = {
                    count = 0
                    interval = 20
                }

	}
	name = "updated name for synthetics browser test bar"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "buz"]
	status = "live"
}
`

func testSyntheticsTestExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := accProvider.Meta().(*datadog.Client)

		for _, r := range s.RootModule().Resources {
			if _, err := client.GetSyntheticsTest(r.Primary.ID); err != nil {
				return fmt.Errorf("Received an error retrieving synthetics test %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsTestIsDestroyed(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := accProvider.Meta().(*datadog.Client)

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
}
