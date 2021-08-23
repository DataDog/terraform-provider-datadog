package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogSyntheticsAPITest_importBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	testName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsAPITestConfig(testName),
			},
			{
				ResourceName:      "datadog_synthetics_test.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsSSLTest_importBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	testName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsSSLTestConfig(testName),
			},
			{
				ResourceName:      "datadog_synthetics_test.ssl",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsTCPTest_importBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	testName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	testName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	testName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsBrowserTestConfig(testName),
			},
			{
				ResourceName:            "datadog_synthetics_test.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"options_list", "browser_variable", "browser_step"},
			},
		},
	})
}

func TestAccDatadogSyntheticsAPITest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsAPITest_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStep(ctx, accProvider, t),
			updateSyntheticsAPITestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsAPITest_BasicNewAssertionsOptions(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStepNewAssertionsOptions(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsAPITest_UpdatedNewAssertionsOptions(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStepNewAssertionsOptions(ctx, accProvider, t),
			updateSyntheticsAPITestStepNewAssertionsOptions(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsSSLTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsSSLTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsSSLTest_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsSSLTestStep(ctx, accProvider, t),
			updateSyntheticsSSLTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsSSLMissingTagsAttributeTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsSSLMissingTagsAttributeTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsTCPTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsTCPTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsTCPTest_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsTCPTestStep(ctx, accProvider, t),
			updateSyntheticsTCPTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsDNSTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsDNSTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsDNSTest_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsDNSTestStep(ctx, accProvider, t),
			updateSyntheticsDNSTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsICMPTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsICMPTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsBrowserTest_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStep(ctx, accProvider, t),
			updateSyntheticsBrowserTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsBrowserTestBrowserVariables_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestBrowserVariablesStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsBrowserTestBrowserNewBrowserStep_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	testName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStepNewBrowserStep(ctx, accProvider, t, testName),
		},
	})
}

func TestAccDatadogSyntheticsTestBrowserMML_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	testName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStepMML(ctx, accProvider, t, testName),
			updateBrowserTestMML(ctx, accProvider, t, testName),
			updateSyntheticsBrowserTestMmlStep(ctx, accProvider, t),
			updateSyntheticsBrowserTestForceMmlStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsTestMultistepApi_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsMultistepAPITest(ctx, accProvider, t),
		},
	})
}

func createSyntheticsAPITestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsAPITestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "subtype", "http"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_definition.0.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_definition.0.timeout", "30"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_definition.0.body", "this is a body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_definition.0.no_saving_response_body", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.#", "4"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.0.type", "header"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.0.property", "content-type"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.0.operator", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.0.target", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.1.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.1.target", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.2.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.2.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.2.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.3.type", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.3.operator", "doesNotContain"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.3.target", "terraform"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "locations.#", "1"),
			resource.TestCheckTypeSetElemAttr(
				"datadog_synthetics_test.foo", "locations.*", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.allow_insecure", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.follow_redirects", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.retry.0.count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.monitor_name", fmt.Sprintf(`%s-monitor`, testName)),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.monitor_priority", "5"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0_list.#", "0"),
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
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "config_variable.0.type", "text"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "config_variable.0.name", "VARIABLE_NAME"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "config_variable.0.pattern", "{{numeric(3)}}"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "config_variable.0.example", "123"),
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

	request_definition {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		timeout = 30
		no_saving_response_body = true
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
		type = "responseTime"
		operator = "lessThan"
		target = "2000"
	}
	assertion {
		type = "body"
		operator = "doesNotContain"
		target = "terraform"
	}

	locations = [ "aws:eu-central-1" ]

	options_list {
		allow_insecure = true
		tick_every = 60
		follow_redirects = true
		min_failure_duration = 0
		min_location_failed = 1
		retry {
			count = 1
		}
		monitor_name = "%[1]s-monitor"
		monitor_priority = 5
	}

	name = "%[1]s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"

	config_variable {
		type = "text"
		name = "VARIABLE_NAME"
		pattern = "{{numeric(3)}}"
		example = "123"
	}
}`, uniq)
}

func createSyntheticsAPITestStepNewAssertionsOptions(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsAPITestConfigNewAssertionsOptions(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "subtype", "http"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.url", "https://www.datadoghq.com"),
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
				"datadog_synthetics_test.bar", "request_client_certificate.0.cert.0.content", utils.ConvertToSha256("content-certificate")),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_client_certificate.0.cert.0.filename", "Provided in Terraform config"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_client_certificate.0.key.0.content", utils.ConvertToSha256("content-key")),
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

	request_definition {
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

func updateSyntheticsAPITestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsAPITestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "subtype", "http"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_definition.0.url", "https://docs.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_definition.0.timeout", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.0.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.0.operator", "isNot"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "assertion.0.target", "500"),
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

	request_definition {
		method = "GET"
		url = "https://docs.datadoghq.com"
		timeout = 60
	}

	assertion {
		type = "statusCode"
		operator = "isNot"
		target = "500"
	}

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

func updateSyntheticsAPITestStepNewAssertionsOptions(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "updated"
	return resource.TestStep{
		Config: updateSyntheticsAPITestConfigNewAssertionsOptions(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "subtype", "http"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.url", "https://docs.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.timeout", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_client_certificate.0.cert.0.content", utils.ConvertToSha256("content-certificate-updated")),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_client_certificate.0.cert.0.filename", "Provided in Terraform config"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_client_certificate.0.key.0.content", utils.ConvertToSha256("content-key-updated")),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_client_certificate.0.key.0.filename", "key-updated"),
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

	request_definition {
		method = "GET"
		url = "https://docs.datadoghq.com"
		timeout = 60
	}

	request_client_certificate {
		cert {
			content = "content-certificate-updated"
		}
		key {
			content = "content-key-updated"
			filename = "key-updated"
		}
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

func createSyntheticsSSLTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsSSLTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "subtype", "ssl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request_definition.0.host", "datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request_definition.0.port", "443"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.0.type", "certificate"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.0.operator", "isInMoreThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.0.target", "30"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options_list.0.accept_self_signed", "true"),
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

	request_definition {
		host = "datadoghq.com"
		port = 443
	}

	assertion {
		type = "certificate"
		operator = "isInMoreThan"
		target = 30
	}

	locations = [ "aws:eu-central-1" ]
	options_list {
		tick_every = 60
		accept_self_signed = true
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = []

	status = "paused"
}`, uniq)
}

func createSyntheticsSSLMissingTagsAttributeTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsSSLMissingTagsAttributeTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "subtype", "ssl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request_definition.0.host", "datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request_definition.0.port", "443"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.0.type", "certificate"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.0.operator", "isInMoreThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.0.target", "30"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options_list.0.accept_self_signed", "true"),
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

	request_definition {
		host = "datadoghq.com"
		port = 443
	}

	assertion {
		type = "certificate"
		operator = "isInMoreThan"
		target = 30
	}

	locations = [ "aws:eu-central-1" ]
	options_list {
		tick_every = 60
		accept_self_signed = true
	}

	name = "%s"
	message = "Notify @datadog.user"

	status = "paused"
}`, uniq)
}

func updateSyntheticsSSLTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsSSLTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "subtype", "ssl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request_definition.0.host", "datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "request_definition.0.port", "443"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.0.type", "certificate"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.0.operator", "isInMoreThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "assertion.0.target", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "locations.0", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.ssl", "options_list.0.accept_self_signed", "false"),
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

	request_definition {
		host = "datadoghq.com"
		port = 443
	}

	assertion {
		type = "certificate"
		operator = "isInMoreThan"
		target = 60
	}

	locations = [ "aws:eu-central-1" ]

	options_list {
		tick_every = 60
		accept_self_signed = false
	}

	name = "%s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"
}`, uniq)
}

func createSyntheticsTCPTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsTCPTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "subtype", "tcp"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "request_definition.0.host", "agent-intake.logs.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "request_definition.0.port", "443"),
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

	request_definition {
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

func updateSyntheticsTCPTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsTCPTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "subtype", "tcp"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "request_definition.0.host", "agent-intake.logs.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "request_definition.0.port", "443"),
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

	request_definition {
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

func createSyntheticsDNSTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsDNSTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "subtype", "dns"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "request_definition.0.host", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "request_definition.0.dns_server", "8.8.8.8"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "request_definition.0.dns_server_port", "120"),
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

	request_definition {
		host = "https://www.datadoghq.com"
		dns_server = "8.8.8.8"
		dns_server_port = 120
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

func updateSyntheticsDNSTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsDNSTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "subtype", "dns"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "request_definition.0.host", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.dns", "request_definition.0.dns_server", "8.8.8.8"),
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

	request_definition {
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

func createSyntheticsICMPTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsICMPTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "subtype", "icmp"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "request_definition.0.host", "www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "request_definition.0.number_of_packets", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "request_definition.0.should_track_hops", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.type", "latency"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.property", "avg"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.target", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "locations.#", "1"),
			resource.TestCheckTypeSetElemAttr(
				"datadog_synthetics_test.icmp", "locations.*", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.icmp", "monitor_id"),
		),
	}
}

func createSyntheticsICMPTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "icmp" {
	type = "api"
	subtype = "icmp"

	request_definition {
		host = "www.datadoghq.com"
		number_of_packets = 2
		should_track_hops = true
	}

	assertion {
		type = "latency"
		operator = "lessThan"
		property = "avg"
		target = 200
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

func createSyntheticsBrowserTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsBrowserTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "browser"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.body", "this is a body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.timeout", "30"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_headers.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_headers.Accept", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_headers.X-Datadog-Trace-ID", "123456789"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "set_cookie", "name=value"),
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
				"datadog_synthetics_test.bar", "options_list.0.no_screenshot", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.monitor_name", fmt.Sprintf(`%s-monitor`, testName)),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.monitor_priority", "5"),
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
				"datadog_synthetics_test.bar", "browser_step.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.name", "first step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.type", "assertCurrentUrl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.check", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.value", "content"),
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
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "config_variable.0.type", "text"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "config_variable.0.name", "VARIABLE_NAME"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "config_variable.0.pattern", "{{numeric(3)}}"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "config_variable.0.example", "123"),
		),
	}
}

func createSyntheticsBrowserTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
	type = "browser"

	request_definition {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		timeout = 30
	}
	request_headers = {
		Accept = "application/json"
		X-Datadog-Trace-ID = "123456789"
	}

	set_cookie = "name=value"

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
		monitor_name = "%[1]s-monitor"
		monitor_priority = 5

		no_screenshot = true
	}

	name = "%[1]s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"

	browser_step {
	    name = "first step"
	    type = "assertCurrentUrl"
	    params {
	        check = "contains"
	        value = "content"
	    }
	}

	browser_variable {
		type = "text"
		name = "MY_PATTERN_VAR"
		pattern = "{{numeric(3)}}"
		example = "597"
	}

	config_variable {
		type = "text"
		name = "VARIABLE_NAME"
		pattern = "{{numeric(3)}}"
		example = "123"
	}
}`, uniq)
}

func updateSyntheticsBrowserTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated"
	return resource.TestStep{
		Config: updateSyntheticsBrowserTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "browser"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.method", "PUT"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.url", "https://docs.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.body", "this is an updated body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.timeout", "60"),
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
				"datadog_synthetics_test.bar", "options_list.0.no_screenshot", "false"),
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
				"datadog_synthetics_test.bar", "browser_step.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.name", "first step updated"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.type", "assertCurrentUrl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.check", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.value", "content"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.1.name", "press key step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.1.type", "pressKey"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.1.params.0.value", "1"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.0.type", "text"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.0.name", "MY_PATTERN_VAR"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.0.pattern", "{{numeric(4)}}"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.0.example", "5970"),
		),
	}
}

func updateSyntheticsBrowserTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
	type = "browser"
	request_definition {
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

		no_screenshot = false
	}
	name = "%s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "buz"]
	status = "live"

	browser_step {
	    name = "first step updated"
	    type = "assertCurrentUrl"

	    params {
	        check = "contains"
	        value = "content"
	    }
	}

	browser_step {
	    name = "press key step"
	    type = "pressKey"

	    params {
	        value = "1"
	    }
	}

	browser_variable {
		type = "text"
		name = "MY_PATTERN_VAR"
		pattern = "{{numeric(4)}}"
		example = "5970"
	}
}`, uniq)
}

func createSyntheticsBrowserTestBrowserVariablesStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsBrowserTestBrowserVariablesConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "browser"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.url", "https://www.datadoghq.com"),
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
				"datadog_synthetics_test.bar", "browser_step.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.name", "first step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.type", "assertCurrentUrl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.check", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.value", "content"),
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

       request_definition {
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

       browser_step {
           name = "first step"
           type = "assertCurrentUrl"
           params {
               check = "contains"
               value = "content"
           }
       }

       browser_variable {
               type = "text"
               name = "MY_PATTERN_VAR"
               pattern = "{{numeric(3)}}"
               example = "597"
       }
}`, uniq)
}

func createSyntheticsBrowserTestStepNewBrowserStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T, testName string) resource.TestStep {
	return resource.TestStep{
		Config: createSyntheticsBrowserTestNewBrowserStepConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "browser"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.body", "this is a body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.timeout", "30"),
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
				"datadog_synthetics_test.bar", "browser_step.#", "7"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.name", "first step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.type", "assertCurrentUrl"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.check", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.value", "content"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.1.name", "scroll step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.1.type", "scroll"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.1.params.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.1.params.0.x", "100"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.1.params.0.y", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.2.name", "api step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.2.type", "runApiTest"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.2.params.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.2.params.0.request", "{\"config\":{\"assertions\":[],\"request\":{\"method\":\"GET\",\"url\":\"https://example.com\"}},\"options\":{},\"subtype\":\"http\"}"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.3.name", "subtest"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.3.type", "playSubTest"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.3.params.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.3.params.0.playing_tab_id", "0"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "browser_step.3.params.0.subtest_public_id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.4.name", "wait step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.4.type", "wait"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.4.params.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.4.params.0.value", "100"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.5.name", "extract variable step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.5.type", "extractFromJavascript"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.5.params.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.5.params.0.code", "return 123"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.6.name", "click step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.6.type", "click"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.6.params.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.6.params.0.element", MML),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
		),
	}
}

func createSyntheticsBrowserTestNewBrowserStepConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "subtest" {
	type = "browser"

	request_definition {
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

		monitor_options {
			renotify_interval = 100
		}
	}

	name = "%[1]s-subtest"
	message = ""
	tags = []

	status = "paused"
}

resource "datadog_synthetics_test" "bar" {
	type = "browser"

	request_definition {
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

	name = "%[1]s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"

	browser_step {
	    name = "first step"
	    type = "assertCurrentUrl"
	    params {
	    	check = "contains"
	    	value = "content"
	    }
	}

	browser_step {
	    name = "scroll step"
	    type = "scroll"
	    params {
	    	x = 100
	    	y = 200
	    }
	}

	browser_step {
		name = "api step"
		type = "runApiTest"
		params {
			request = jsonencode({
				"config": {
					"assertions": [],
					"request": {
						"method": "GET",
						"url": "https://example.com"
					}
				},
				"options": {}
        		"subtype": "http",
      		})
		}
	}

	browser_step {
		name = "subtest"
		type = "playSubTest"
		params {
			playing_tab_id = 0
			subtest_public_id = datadog_synthetics_test.subtest.id
		}
	}

	browser_step {
		name = "wait step"
		type = "wait"
		params {
			value = "100"
		}
	}

	browser_step {
		name = "extract variable step"
		type = "extractFromJavascript"
		params {
			code = "return 123"
			variable {
				name = "VAR_FROM_JS"
			}
		}
	}

	browser_step {
		name = "click step"
		type = "click"
		params {
			element = jsonencode({
				"multiLocator": {
					"ab": "/*[local-name()=\"html\"][1]/*[local-name()=\"body\"][1]/*[local-name()=\"nav\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"a\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"img\"][1]",
					"at": "/descendant::*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]",
					"cl": "/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]",
					"clt": "/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]",
					"co": "",
					"ro": "//*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]"
				},
				"targetOuterHTML": "img height=\"75\" src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png...",
				"url": "https://www.datadoghq.com/"
			})
		}
	}
}`, uniq)
}

const MML = `{"multiLocator":{"ab":"/*[local-name()=\"html\"][1]/*[local-name()=\"body\"][1]/*[local-name()=\"nav\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"a\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"img\"][1]","at":"/descendant::*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]","cl":"/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]","clt":"/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]","co":"","ro":"//*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]"},"targetOuterHTML":"img height=\"75\" src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png...","url":"https://www.datadoghq.com/"}`

const MMLManualUpdate = `{"multiLocator":{"ab":"/*[local-name()=\"html\"][1]/*[local-name()=\"body\"][1]/*[local-name()=\"nav\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"a\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"img\"][1]","at":"/descendant::*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]","cl":"/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]","clt":"/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]","co":"","ro":"//*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]"},"targetOuterHTML":"img height=\"75\" src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png...","url":"https://www.datadoghq.com/updated"}`

const MMLConfigUpdate = `{"multiLocator":{"ab":"/*[local-name()=\"html\"][1]/*[local-name()=\"body\"][1]/*[local-name()=\"nav\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"a\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"img\"][1]","at":"/descendant::*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]","cl":"/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]","clt":"/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]","co":"","ro":"//*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]"},"targetOuterHTML":"img height=\"75\" src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png...","url":"https://www.datadoghq.com/config-updated"}`

func createSyntheticsBrowserTestStepMML(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T, testName string) resource.TestStep {
	return resource.TestStep{
		Config: createSyntheticsBrowserTestMMLConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "browser"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.body", "this is a body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.timeout", "30"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_headers.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_headers.Accept", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_headers.X-Datadog-Trace-ID", "123456789"),
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
				"datadog_synthetics_test.bar", "browser_step.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.name", "click step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.type", "click"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.element", MML),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.bar", "monitor_id"),
		),
	}
}

func createSyntheticsBrowserTestMMLConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
	type = "browser"

	request_definition {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		timeout = 30
	}
	request_headers = {
		Accept = "application/json"
		X-Datadog-Trace-ID = "123456789"
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

		monitor_options {
			renotify_interval = 100
		}
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"

	browser_step {
		name = "click step"
		type = "click"
		params {
			element = jsonencode({
				"multiLocator": {
					"ab": "/*[local-name()=\"html\"][1]/*[local-name()=\"body\"][1]/*[local-name()=\"nav\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"a\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"img\"][1]",
					"at": "/descendant::*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]",
					"cl": "/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]",
					"clt": "/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]",
					"co": "",
					"ro": "//*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]"
				},
				"targetOuterHTML": "img height=\"75\" src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png...",
				"url": "https://www.datadoghq.com/"
			})
		}
	}
}`, uniq)
}

func updateBrowserTestMML(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T, testName string) resource.TestStep {
	return resource.TestStep{
		Config: createSyntheticsBrowserTestMMLConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			editSyntheticsTestMML(accProvider),
		),
	}
}

func updateSyntheticsBrowserTestMmlStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated"
	return resource.TestStep{
		Config: createSyntheticsBrowserTestMMLConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "browser"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "device_ids.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "device_ids.0", "laptop_large"),
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
				"datadog_synthetics_test.bar", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "status", "paused"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.name", "click step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.type", "click"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.element", MMLManualUpdate),
		),
	}
}

func updateSyntheticsBrowserTestForceMmlStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated"
	return resource.TestStep{
		Config: createSyntheticsBrowserForceMmlTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "type", "browser"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "device_ids.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "device_ids.0", "laptop_large"),
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
				"datadog_synthetics_test.bar", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "status", "paused"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.name", "click step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.type", "click"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.element", MMLConfigUpdate),
		),
	}
}

func createSyntheticsBrowserForceMmlTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
	type = "browser"

	request_definition {
		method = "GET"
		url = "https://www.datadoghq.com"
	}

	device_ids = [ "laptop_large" ]
	locations = [ "aws:eu-central-1" ]
	options_list {
		tick_every = 900
		min_failure_duration = 0
		min_location_failed = 1
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"

	browser_step {
		name = "click step"
		type = "click"
		force_element_update = true
		params {
			element = jsonencode({
				"multiLocator": {
					"ab": "/*[local-name()=\"html\"][1]/*[local-name()=\"body\"][1]/*[local-name()=\"nav\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"a\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"img\"][1]",
					"at": "/descendant::*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]",
					"cl": "/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]",
					"clt": "/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]",
					"co": "",
					"ro": "//*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]"
				},
				"targetOuterHTML": "img height=\"75\" src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png...",
				"url": "https://www.datadoghq.com/config-updated"
			})
		}
	}
}`, uniq)
}

func createSyntheticsMultistepAPITest(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	variableName := getUniqueVariableName(ctx, t)

	return resource.TestStep{
		Config: createSyntheticsMultistepAPITestConfig(testName, variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "subtype", "multi"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "locations.#", "1"),
			resource.TestCheckTypeSetElemAttr(
				"datadog_synthetics_test.multi", "locations.*", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "options_list.0.tick_every", "900"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "options_list.0.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "options_list.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "tags.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "tags.0", "multistep"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "status", "paused"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.name", "First api step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_definition.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_definition.0.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_definition.0.body", "this is a body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_definition.0.timeout", "30"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_definition.0.allow_insecure", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_headers.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_headers.Accept", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_headers.X-Datadog-Trace-ID", "123456789"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_query.%", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_query.foo", "bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_basicauth.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_basicauth.0.username", "admin"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_basicauth.0.password", "secret"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_client_certificate.0.cert.0.filename", "Provided in Terraform config"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_client_certificate.0.cert.0.content", utils.ConvertToSha256("content-certificate")),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_client_certificate.0.key.0.filename", "key"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_client_certificate.0.key.0.content", utils.ConvertToSha256("content-key")),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.assertion.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.assertion.0.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.assertion.0.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.assertion.0.target", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.extracted_value.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.extracted_value.0.name", "VAR_EXTRACT"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.extracted_value.0.type", "http_header"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.extracted_value.0.field", "content-length"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.extracted_value.0.parser.0.type", "regex"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.extracted_value.0.parser.0.value", ".*"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.allow_failure", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.is_critical", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "config_variable.0.type", "global"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "config_variable.0.name", "VARIABLE_NAME"),
		),
	}
}

func createSyntheticsMultistepAPITestConfig(testName string, variableName string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "global_variable" {
	name = "%[2]s"
	description = "a global variable"
	tags = ["foo:bar", "baz"]
	value = "variable-value"
}

resource "datadog_synthetics_test" "multi" {
       type = "api"
       subtype = "multi"
       locations = ["aws:eu-central-1"]
       options_list {
               tick_every = 900
               min_failure_duration = 0
               min_location_failed = 1
       }
       name = "%[1]s"
       message = "Notify @datadog.user"
       tags = ["multistep"]
       status = "paused"

       config_variable {
       	   id = datadog_synthetics_global_variable.global_variable.id
       	   type = "global"
           name = "VARIABLE_NAME"
       }

       api_step {
               name = "First api step"
               request_definition {
                       method = "GET"
                       url = "https://www.datadoghq.com"
                       body = "this is a body"
                       timeout = 30
                       allow_insecure = true
               }
               request_headers = {
               	       Accept = "application/json"
               	       X-Datadog-Trace-ID = "123456789"
               }
               request_query = {
                       foo = "bar"
               }
               request_basicauth {
                       username = "admin"
               	       password = "secret"
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
                       type = "statusCode"
                       operator = "is"
                       target = "200"
               }

               extracted_value {
               		   name = "VAR_EXTRACT"
               		   field = "content-length"
               		   type = "http_header"
               		   parser {
               		   		   type = "regex"
               		   		   value = ".*"
               		   }
               }
               allow_failure = true
               is_critical = false
       }
}
`, testName, variableName)
}

func testSyntheticsTestExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, _, err := datadogClientV1.SyntheticsApi.GetTest(authV1, r.Primary.ID); err != nil {
				return fmt.Errorf("received an error retrieving synthetics test %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsTestIsDestroyed(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, _, err := datadogClientV1.SyntheticsApi.GetTest(authV1, r.Primary.ID); err != nil {
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

func editSyntheticsTestMML(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			provider, _ := accProvider()
			providerConf := provider.Meta().(*datadog.ProviderConfiguration)
			datadogClientV1 := providerConf.DatadogClientV1
			authV1 := providerConf.AuthV1

			syntheticsTest, _, err := datadogClientV1.SyntheticsApi.GetBrowserTest(authV1, r.Primary.ID)

			if err != nil {
				return fmt.Errorf("failed to read synthetics test %s", err)
			}

			syntheticsTestUpdate := datadogV1.NewSyntheticsBrowserTest(syntheticsTest.GetMessage())
			syntheticsTestUpdate.SetName(syntheticsTest.GetName())
			syntheticsTestUpdate.SetType(datadogV1.SYNTHETICSBROWSERTESTTYPE_BROWSER)
			syntheticsTestUpdate.SetConfig(syntheticsTest.GetConfig())
			syntheticsTestUpdate.SetStatus(syntheticsTest.GetStatus())
			syntheticsTestUpdate.SetLocations(syntheticsTest.GetLocations())
			syntheticsTestUpdate.SetOptions(syntheticsTest.GetOptions())
			syntheticsTestUpdate.SetTags(syntheticsTest.GetTags())

			// manually update the MML so the state is outdated
			step := datadogV1.SyntheticsStep{}
			step.SetName("click step")
			step.SetType(datadogV1.SYNTHETICSSTEPTYPE_CLICK)
			params := make(map[string]interface{})
			elementParams := `{"element":` + MMLManualUpdate + "}"
			utils.GetMetadataFromJSON([]byte(elementParams), &params)
			step.SetParams(params)
			steps := []datadogV1.SyntheticsStep{step}
			syntheticsTestUpdate.SetSteps(steps)

			if _, _, err := datadogClientV1.SyntheticsApi.UpdateBrowserTest(authV1, r.Primary.ID, *syntheticsTestUpdate); err != nil {
				return fmt.Errorf("failed to manually update synthetics test %s", err)
			}
		}

		return nil
	}
}
