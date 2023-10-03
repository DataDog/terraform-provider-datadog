package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogSyntheticsAPITest_importBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	testName := uniqueEntityName(ctx, t)
	variableName := getUniqueVariableName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsAPITestConfig(testName, variableName),
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

func TestAccDatadogSyntheticsAPITest_AdvancedScheduling(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsAPITestStepAdvancedScheduling(ctx, accProvider, t),
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

func TestAccDatadogSyntheticsUDPTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsUDPTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsWebsocketTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsWebsocketTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGRPCTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGRPCTestStep(ctx, accProvider, t),
			updateSyntheticsGRPCTestStep(ctx, accProvider, t),
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

func TestAccDatadogSyntheticsBrowserTest_Updated_RumSettings(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStep(ctx, accProvider, t),
			updateSyntheticsBrowserTestStepRumSettings(ctx, accProvider, t),
			updateSyntheticsBrowserTestStepRumSettingsEnabled(ctx, accProvider, t),
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

func TestAccDatadogSyntheticsTestBrowserUserLocator_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	testName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStepUserLocator(ctx, accProvider, t, testName),
		},
	})
}

func TestAccDatadogSyntheticsTestBrowserUserLocator_NoElement(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	testName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsBrowserTestStepUserLocatorNoElement(ctx, accProvider, t, testName),
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
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsAPITestConfig(testName, variableName),
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
				"datadog_synthetics_test.foo", "request_definition.0.body_type", "text/plain"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_definition.0.no_saving_response_body", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_basicauth.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_basicauth.0.type", "ntlm"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_basicauth.0.username", "ntlm-username"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_basicauth.0.password", "ntlm-password"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_basicauth.0.domain", "ntlm-domain"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_basicauth.0.workstation", "ntlm-workstation"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_proxy.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_proxy.0.url", "https://proxy.url"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_proxy.0.headers.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_proxy.0.headers.Accept", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "request_proxy.0.headers.X-Datadog-Trace-ID", "123456789"),
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
				"datadog_synthetics_test.foo", "assertion.2.timings_scope", "withoutDNS"),
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
				"datadog_synthetics_test.foo", "options_list.0.restricted_roles.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.ci.0.execution_rule", "blocking"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.ignore_server_certificate_error", "true"),
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
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "config_variable.0.secure", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "config_variable.1.type", "global"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "config_variable.1.name", "GLOBAL_VAR"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.foo", "monitor_id"),
		),
	}
}

func createSyntheticsAPITestConfig(uniq string, variableName string) string {
	return fmt.Sprintf(`
resource "datadog_role" "bar" {
	name      = "%[1]s"
}

resource "datadog_synthetics_global_variable" "global_variable" {
  name        = "%[2]s"
  description = "a global variable"
  tags        = ["foo:bar", "baz"]
  value       = "variable-value"
}

resource "datadog_synthetics_test" "foo" {
	type = "api"
	subtype = "http"

	request_definition {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		body_type = "text/plain"
		timeout = 30
		no_saving_response_body = true
	}
	request_headers = {
		Accept = "application/json"
		X-Datadog-Trace-ID = "123456789"
	}

	request_basicauth {
		type = "ntlm"
		username = "ntlm-username"
		password = "ntlm-password"
		domain = "ntlm-domain"
		workstation = "ntlm-workstation"
	}

	request_proxy {
		url = "https://proxy.url"
		headers = {
			Accept = "application/json"
			X-Datadog-Trace-ID = "123456789"
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
		type = "responseTime"
		operator = "lessThan"
		target = "2000"
		timings_scope = "withoutDNS"
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
		http_version = "http2"
		retry {
			count = 1
		}
		monitor_name = "%[1]s-monitor"
		monitor_priority = 5
		restricted_roles = ["${datadog_role.bar.id}"]
		ci {
			execution_rule = "blocking"
		}
		ignore_server_certificate_error = true
	}

	name = "%[1]s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"

	config_variable {
		type = "text"
		name = "VARIABLE_NAME"
		secure = false
		pattern = "{{numeric(3)}}"
		example = "123"
	}

	config_variable {
		type = "global"
		name = "GLOBAL_VAR"
		id   = datadog_synthetics_global_variable.global_variable.id
		secure = false
	}
}`, uniq, variableName)
}

func createSyntheticsAPITestConfigAdvancedScheduling(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "advanced_scheduling" {
	type = "api"
	subtype = "http"

	request_definition {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		body_type = "text/plain"
		timeout = 30
		no_saving_response_body = true
		persist_cookies = true
	}
	request_headers = {
		Accept = "application/json"
		X-Datadog-Trace-ID = "123456789"
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

	locations = [ "aws:eu-central-1" ]

	options_list {
		allow_insecure = true
		tick_every = 60
		follow_redirects = true
		min_failure_duration = 0
		min_location_failed = 1
		http_version = "http2"
		retry {
			count = 1
		}
		monitor_name = "%[1]s-monitor"
		monitor_priority = 5
		ci {
			execution_rule = "blocking"
		}
		ignore_server_certificate_error = true
		scheduling {
			timeframes {
				day=1
				from="07:00"
				to="18:00"
			}
			timezone = "America/New_York"
		}
	}

	name = "%[1]s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}`, uniq)
}

func createSyntheticsAPITestStepAdvancedScheduling(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsAPITestConfigAdvancedScheduling(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "subtype", "http"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "request_definition.0.method", "GET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "request_definition.0.url", "https://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "request_definition.0.timeout", "30"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "request_definition.0.body", "this is a body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "request_definition.0.body_type", "text/plain"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "request_definition.0.no_saving_response_body", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "request_definition.0.persist_cookies", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.0.type", "header"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.0.property", "content-type"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.0.operator", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.0.target", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.1.type", "statusCode"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.1.target", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.2.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.2.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "assertion.2.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "locations.#", "1"),
			resource.TestCheckTypeSetElemAttr(
				"datadog_synthetics_test.advanced_scheduling", "locations.*", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.allow_insecure", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.follow_redirects", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.min_failure_duration", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.min_location_failed", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.retry.0.count", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.monitor_name", fmt.Sprintf(`%s-monitor`, testName)),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.monitor_priority", "5"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0_list.#", "0"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.ci.0.execution_rule", "blocking"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.ignore_server_certificate_error", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.scheduling.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.scheduling.0.timeframes.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.scheduling.0.timeframes.0.from", "07:00"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.scheduling.0.timeframes.0.to", "18:00"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.scheduling.0.timeframes.0.day", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "options_list.0.scheduling.0.timezone", "America/New_York"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.advanced_scheduling", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.advanced_scheduling", "monitor_id"),
		),
	}
}

func createSyntheticsAPITestStepNewAssertionsOptions(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	globalVariableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsAPITestConfigNewAssertionsOptions(testName, globalVariableName),
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
				"datadog_synthetics_test.bar", "request_basicauth.0.type", "web"),
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
				"datadog_synthetics_test.bar", "assertion.#", "10"),
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
				"datadog_synthetics_test.bar", "assertion.7.type", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.7.operator", "validatesJSONPath"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.7.targetjsonpath.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.7.targetjsonpath.0.jsonpath", "$.mykey"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.7.targetjsonpath.0.operator", "moreThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.7.targetjsonpath.0.targetvalue", fmt.Sprintf("{{ %s }}", globalVariableName)),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.8.type", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.8.operator", "validatesXPath"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.8.targetxpath.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.8.targetxpath.0.xpath", "something"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.8.targetxpath.0.operator", "contains"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.8.targetxpath.0.targetvalue", "12"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.9.type", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.9.operator", "validatesJSONPath"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.9.targetjsonpath.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.9.targetjsonpath.0.jsonpath", "$.myKey"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "assertion.9.targetjsonpath.0.operator", "isUndefined"),
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
				"datadog_synthetics_test.bar", "options_list.0.monitor_options.0.renotify_interval", "120"),
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

func createSyntheticsAPITestConfigNewAssertionsOptions(uniq, globalVariableName string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "test_variable" {
	name        = "%s"
	description = "Description of the variable"
	tags        = ["foo:bar", "env:test"]
	value       = "variable-value"
}

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
		X-Datadog-Trace-ID = "123456789"
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

	config_variable {
		type = "text"
		name = "TEST"
		example = "1234"
		pattern = "{{ numeric(4) }}"
	}

	config_variable {
		id = datadog_synthetics_global_variable.test_variable.id
		name = datadog_synthetics_global_variable.test_variable.name
		secure = "false"
		type = "global"
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

	assertion {
		type     = "body"
		operator = "validatesJSONPath"
		targetjsonpath {
			jsonpath    = "$.mykey"
			operator    = "moreThan"
			targetvalue = "{{ ${datadog_synthetics_global_variable.test_variable.name} }}"
		}
	}
	assertion {
		type = "body"
		operator = "validatesXPath"
		targetxpath {
			operator = "contains"
			targetvalue = "12"
			xpath = "something"
        }
    }
    assertion {
     	operator = "validatesJSONPath"
		type     = "body"
		targetjsonpath {
			jsonpath    = "$.myKey"
			operator    = "isUndefined"
		}
    }

	locations = [ "aws:eu-central-1" ]
	options_list {
		tick_every = 60
		follow_redirects = true
		min_failure_duration = 0
		min_location_failed = 1

		monitor_options {
			renotify_interval = 120
		}
	}

	name = "%s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"
}`, globalVariableName, uniq)
}

func updateSyntheticsAPITestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated"
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: updateSyntheticsAPITestConfig(testName, variableName),
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
				"datadog_synthetics_test.foo", "options_list.0.monitor_options.0.renotify_interval", "120"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.foo", "options_list.0.ci.0.execution_rule", "non_blocking"),
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

func updateSyntheticsAPITestConfig(uniq string, varName string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "global_variable" {
	name        = "%[2]s"
	description = "a global variable"
	tags        = ["foo:bar", "baz"]
	value       = "variable-value"
}

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
			renotify_interval = 120
		}
		ci {
			execution_rule = "non_blocking"
		}
	}

	name = "%[1]s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"

	config_variable {
		type = "global"
		name = "GLOBAL_VAR"
		id   = datadog_synthetics_global_variable.global_variable.id
		secure = false
	}
}`, uniq, varName)
}

func updateSyntheticsAPITestStepNewAssertionsOptions(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "updated"
	globalVariableName := getUniqueVariableName(ctx, t)

	return resource.TestStep{
		Config: updateSyntheticsAPITestConfigNewAssertionsOptions(testName, globalVariableName),
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

func updateSyntheticsAPITestConfigNewAssertionsOptions(uniq, globalVariableName string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "test_variable" {
	name        = "%s"
	description = "Description of the variable"
	tags        = ["foo:bar", "env:test"]
	value       = "variable-value"
}

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

	config_variable {
		id = datadog_synthetics_global_variable.test_variable.id
		name = datadog_synthetics_global_variable.test_variable.name
		secure = "false"
		type = "global"
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
}`, globalVariableName, uniq)
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
				"datadog_synthetics_test.ssl", "request_definition.0.servername", "datadoghq.com"),
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
				"datadog_synthetics_test.ssl", "options_list.0.check_certificate_revocation", "true"),
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
		servername = "datadoghq.com"
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
		check_certificate_revocation = true
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
				"datadog_synthetics_test.ssl", "options_list.0.check_certificate_revocation", "false"),
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
		check_certificate_revocation = false
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
				"datadog_synthetics_test.tcp", "assertion.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.0.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.0.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.0.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.1.type", "connection"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.tcp", "assertion.1.target", "established"),
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

	assertion {
		type = "connection"
		operator = "is"
		target = "established"
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
				"datadog_synthetics_test.icmp", "assertion.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.type", "latency"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.property", "avg"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.target", "200"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.1.type", "packetLossPercentage"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.1.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.1.target", "0.06"),
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

	assertion {
		type = "packetLossPercentage"
		operator = "lessThan"
		target = "0.06"
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

func createSyntheticsUDPTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsUDPTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "subtype", "udp"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "request_definition.0.host", "www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "request_definition.0.message", "message"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "request_definition.0.port", "443"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.0.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.1.type", "receivedMessage"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.icmp", "assertion.1.target", "message"),
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

func createSyntheticsUDPTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "icmp" {
	type = "api"
	subtype = "udp"

	request_definition {
		host = "www.datadoghq.com"
		port = 443
		message = "message"
	}

	assertion {
		type = "responseTime"
		operator = "lessThan"
		target = "2000"
	}
	assertion {
		type = "receivedMessage"
		operator = "is"
		target = "message"
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

func createSyntheticsWebsocketTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsWebsocketTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "subtype", "websocket"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "request_definition.0.url", "wss://www.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "request_definition.0.message", "message"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "assertion.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "assertion.0.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "assertion.0.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "assertion.0.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "assertion.1.type", "receivedMessage"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "assertion.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "assertion.1.target", "message"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "locations.#", "1"),
			resource.TestCheckTypeSetElemAttr(
				"datadog_synthetics_test.websocket", "locations.*", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.websocket", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.websocket", "monitor_id"),
		),
	}
}

func createSyntheticsWebsocketTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "websocket" {
	type = "api"
	subtype = "websocket"

	request_definition {
		url = "wss://www.datadoghq.com"
		message = "message"
	}

	assertion {
		type = "responseTime"
		operator = "lessThan"
		target = "2000"
	}
	assertion {
		type = "receivedMessage"
		operator = "is"
		target = "message"
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

func createSyntheticsGRPCTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGRPCTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "subtype", "grpc"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_definition.0.host", "google.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_definition.0.port", "50050"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_definition.0.service", "Hello"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_metadata.%", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_metadata.header", "value"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.#", "4"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.0.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.0.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.0.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.1.type", "grpcHealthcheckStatus"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.1.target", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.2.type", "grpcProto"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.2.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.2.target", "proto target"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.3.type", "grpcMetadata"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.3.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.3.target", "123"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.3.property", "property"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "locations.#", "1"),
			resource.TestCheckTypeSetElemAttr(
				"datadog_synthetics_test.grpc", "locations.*", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.grpc", "monitor_id"),
		),
	}
}

func createSyntheticsGRPCTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "grpc" {
	type = "api"
	subtype = "grpc"

	request_definition {
		method = "GET"
		host   = "google.com"
		port   = 50050
		service = "Hello"
	}

	request_metadata = {
		header = "value"
	}

	assertion {
		type = "responseTime"
		operator = "lessThan"
		target = "2000"
	}

	assertion {
		operator = "is"
		type     = "grpcHealthcheckStatus"
		target   = 1
	}

	assertion {
		operator = "is"
		target   = "proto target"
		type     = "grpcProto"
	}

	assertion {
		operator = "is"
		target   = "123"
		property = "property"
		type     = "grpcMetadata"
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

func updateSyntheticsGRPCTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: updateSyntheticsGRPCTestConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "type", "api"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "subtype", "grpc"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_definition.0.host", "google.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_definition.0.port", "50050"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_definition.0.service", ""),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_metadata.%", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "request_metadata.header", "value-updated"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.0.type", "responseTime"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.0.operator", "lessThan"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.0.target", "2000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.1.type", "grpcHealthcheckStatus"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.1.operator", "is"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "assertion.1.target", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "locations.#", "1"),
			resource.TestCheckTypeSetElemAttr(
				"datadog_synthetics_test.grpc", "locations.*", "aws:eu-central-1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "options_list.0.tick_every", "60"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "message", "Notify @datadog.user"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.grpc", "status", "paused"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_test.grpc", "monitor_id"),
		),
	}
}

func updateSyntheticsGRPCTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "grpc" {
	type = "api"
	subtype = "grpc"

	request_definition {
		method = "GET"
		host   = "google.com"
		port   = 50050
		service = ""
	}

	request_metadata = {
		header = "value-updated"
	}

	assertion {
		type = "responseTime"
		operator = "lessThan"
		target = "2000"
	}

	assertion {
		operator = "is"
		type     = "grpcHealthcheckStatus"
		target   = 1
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
				"datadog_synthetics_test.bar", "request_definition.0.certificate_domains.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_definition.0.certificate_domains.0", "https://datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_headers.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_headers.Accept", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_headers.X-Datadog-Trace-ID", "123456789"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.0.type", "web"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.0.username", "username"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.0.password", "password"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_proxy.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_proxy.0.url", "https://proxy.url"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_proxy.0.headers.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_proxy.0.headers.Accept", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_proxy.0.headers.X-Datadog-Trace-ID", "123456789"),
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
				"datadog_synthetics_test.bar", "options_list.0.monitor_options.0.renotify_interval", "120"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.no_screenshot", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.monitor_name", fmt.Sprintf(`%s-monitor`, testName)),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.monitor_priority", "5"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.restricted_roles.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.is_enabled", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.application_id", "rum-app-id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.client_token_id", "12345"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.ci.0.execution_rule", "blocking"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.ignore_server_certificate_error", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.disable_csp", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.disable_cors", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.initial_navigation_timeout", "150"),
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
				"datadog_synthetics_test.bar", "browser_step.0.allow_failure", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.is_critical", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.no_screenshot", "true"),
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
				"datadog_synthetics_test.bar", "browser_variable.1.type", "email"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.1.name", "EMAIL_VAR"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.2.type", "text"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.2.name", "MY_SECRET"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_variable.2.secure", "true"),
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
resource "datadog_role" "bar" {
	name      = "%[1]s"
}
resource "datadog_synthetics_test" "bar" {
	type = "browser"

	request_definition {
		method = "GET"
		url = "https://www.datadoghq.com"
		body = "this is a body"
		timeout = 30
		certificate_domains = ["https://datadoghq.com"]
	}
	request_headers = {
		Accept = "application/json"
		X-Datadog-Trace-ID = "123456789"
	}

	request_basicauth {
		username = "username"
		password = "password"
	}

	request_proxy {
		url = "https://proxy.url"
		headers = {
			Accept = "application/json"
			X-Datadog-Trace-ID = "123456789"
		}
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
			renotify_interval = 120
		}
		monitor_name = "%[1]s-monitor"
		monitor_priority = 5
		restricted_roles = ["${datadog_role.bar.id}"]

		no_screenshot = true

		rum_settings {
			is_enabled = true
			application_id = "rum-app-id"
			client_token_id = "12345"
		}

		ci {
			execution_rule = "blocking"
		}

		ignore_server_certificate_error = true
		disable_csp = true
		disable_cors = true
		initial_navigation_timeout = 150
	}

	name = "%[1]s"
	message = "Notify @datadog.user"
	tags = ["foo:bar", "baz"]

	status = "paused"

	browser_step {
	    name = "first step"
	    type = "assertCurrentUrl"
	    allow_failure = true
	    is_critical = true
	    params {
	        check = "contains"
	        value = "content"
	    }
	    no_screenshot = true
	}

	browser_variable {
		type = "text"
		name = "MY_PATTERN_VAR"
		pattern = "{{numeric(3)}}"
		example = "597"
	}

	browser_variable {
		name = "EMAIL_VAR"
		type = "email"
	}

	browser_variable {
		type = "text"
		name = "MY_SECRET"
		pattern = "secret"
		example = "secret"
		secure = true
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
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.is_enabled", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.ci.0.execution_rule", "skipped"),
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

		rum_settings {
			is_enabled = false
		}

		ci {
			execution_rule = "skipped"
		}
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

func updateSyntheticsBrowserTestStepRumSettings(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated-rumsettings"
	return resource.TestStep{
		Config: updateSyntheticsBrowserTestRumSetting(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.is_enabled", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.application_id", ""),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.client_token_id", "0"),
		),
	}
}

func updateSyntheticsBrowserTestRumSetting(uniq string) string {
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

		rum_settings {
			is_enabled = false
			application_id = "rum-app-id-updated"
			client_token_id = "6789"
		}

		ci {
			execution_rule = "skipped"
		}
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

func updateSyntheticsBrowserTestStepRumSettingsEnabled(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	testName := uniqueEntityName(ctx, t) + "-updated-rumsettings"
	return resource.TestStep{
		Config: updateSyntheticsBrowserTestRumSettingEnabled(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.is_enabled", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.application_id", ""),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "options_list.0.rum_settings.0.client_token_id", "0"),
		),
	}
}

func updateSyntheticsBrowserTestRumSettingEnabled(uniq string) string {
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

		rum_settings {
			is_enabled = true
		}

		ci {
			execution_rule = "skipped"
		}
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
				"datadog_synthetics_test.bar", "request_basicauth.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.0.type", "web"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.0.username", "web-username"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "request_basicauth.0.password", "web-password"),
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

       request_basicauth {
		       username = "web-username"
		       password = "web-password"
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
				"datadog_synthetics_test.bar", "options_list.0.monitor_options.0.renotify_interval", "120"),
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
			renotify_interval = 120
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
			renotify_interval = 120
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
				"datadog_synthetics_test.bar", "options_list.0.monitor_options.0.renotify_interval", "120"),
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
			renotify_interval = 120
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

const MMLCustomUserLocator = `{"multiLocator":{"ab":"/*[local-name()=\"html\"][1]/*[local-name()=\"body\"][1]/*[local-name()=\"nav\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"a\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"div\"][1]/*[local-name()=\"img\"][1]","at":"/descendant::*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]","cl":"/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]","clt":"/descendant::*[contains(concat('''', normalize-space(@class), '' ''), \" dog \")]/*[local-name()=\"img\"][1]","co":"","ro":"//*[@src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png\"]"},"targetOuterHTML":"img height=\"75\" src=\"https://imgix.datadoghq.com/img/dd_logo_n_70x75.png...","url":"https://www.datadoghq.com/config-updated","userLocator":{"failTestOnCannotLocate":true,"values":[{"type":"css","value":"user-locator-test"}]}}`

func createSyntheticsBrowserTestStepUserLocator(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T, testName string) resource.TestStep {
	return resource.TestStep{
		Config: createSyntheticsBrowserTestUserLocatorConfig(testName),
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
				"datadog_synthetics_test.bar", "tags.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "status", "paused"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.name", "click step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.type", "click"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.element", MMLCustomUserLocator),
		),
	}
}

const MMLCustomUserLocatorNoElement = `{"userLocator":{"failTestOnCannotLocate":true,"values":[{"type":"css","value":"user-locator-test"}]}}`

func createSyntheticsBrowserTestStepUserLocatorNoElement(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T, testName string) resource.TestStep {
	return resource.TestStep{
		Config: createSyntheticsBrowserTestUserLocatorNoElementConfig(testName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsTestExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "name", testName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.bar", "browser_step.0.params.0.element", MMLCustomUserLocatorNoElement),
		),
	}
}

func createSyntheticsBrowserTestUserLocatorConfig(uniq string) string {
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
	tags = ["foo:bar"]

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
				"url": "https://www.datadoghq.com/config-updated",
			})
			element_user_locator {
				fail_test_on_cannot_locate = true
				value {
					type = "css"
					value = "user-locator-test"
				}
			}
		}
	}
}`, uniq)
}

func createSyntheticsBrowserTestUserLocatorNoElementConfig(uniq string) string {
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
	tags = ["foo:bar"]

	status = "paused"

	browser_step {
		name = "click step"
		type = "click"
		params {
			element_user_locator {
				fail_test_on_cannot_locate = true
				value {
					type = "css"
					value = "user-locator-test"
				}
			}
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
				"datadog_synthetics_test.multi", "options_list.0.restricted_roles.#", "1"),
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
				"datadog_synthetics_test.multi", "api_step.#", "4"),
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
				"datadog_synthetics_test.multi", "api_step.0.request_definition.0.follow_redirects", "true"),
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
				"datadog_synthetics_test.multi", "api_step.0.request_basicauth.0.type", "sigv4"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_basicauth.0.access_key", "sigv4-access-key"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_basicauth.0.secret_key", "sigv4-secret-key"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_basicauth.0.region", "sigv4-region"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_basicauth.0.service_name", "sigv4-service-name"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_basicauth.0.session_token", "sigv4-session-token"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_client_certificate.0.cert.0.filename", "Provided in Terraform config"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_client_certificate.0.cert.0.content", utils.ConvertToSha256("content-certificate")),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_client_certificate.0.key.0.filename", "key"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_client_certificate.0.key.0.content", utils.ConvertToSha256("content-key")),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_proxy.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_proxy.0.url", "https://proxy.url"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_proxy.0.headers.%", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_proxy.0.headers.Accept", "application/json"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.request_proxy.0.headers.X-Datadog-Trace-ID", "123456789"),
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
				"datadog_synthetics_test.multi", "api_step.0.extracted_value.0.secure", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.allow_failure", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.is_critical", "false"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.retry.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.retry.0.count", "5"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.0.retry.0.interval", "1000"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.1.name", "Second api step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.1.request_basicauth.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.1.request_basicauth.0.type", "oauth-client"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.1.request_basicauth.0.audience", "audience"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.1.request_basicauth.0.client_id", "client-id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.1.request_basicauth.0.client_secret", "client-secret"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.1.request_basicauth.0.scope", "scope"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.1.request_basicauth.0.token_api_authentication", "header"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.1.request_basicauth.0.access_token_url", "https://token.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.name", "Third api step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.type", "oauth-rop"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.audience", "audience"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.client_id", "client-id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.client_secret", "client-secret"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.resource", "resource"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.scope", "scope"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.token_api_authentication", "body"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.access_token_url", "https://token.datadoghq.com"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.username", "username"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.2.request_basicauth.0.password", "password"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.3.name", "Fourth api step"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.3.request_basicauth.#", "1"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.3.request_basicauth.0.type", "digest"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.3.request_basicauth.0.username", "username"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "api_step.3.request_basicauth.0.password", "password"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "config_variable.0.type", "global"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_test.multi", "config_variable.0.name", "VARIABLE_NAME"),
		),
	}
}

func createSyntheticsMultistepAPITestConfig(testName string, variableName string) string {
	return fmt.Sprintf(`
resource "datadog_role" "bar" {
	name      = "%[1]s"
}

resource "datadog_synthetics_global_variable" "global_variable" {
  name        = "%[2]s"
  description = "a global variable"
  tags        = ["foo:bar", "baz"]
  value       = "variable-value"
}

resource "datadog_synthetics_test" "multi" {
  type      = "api"
  subtype   = "multi"
  locations = ["aws:eu-central-1"]
  options_list {
    tick_every           = 900
    min_failure_duration = 0
    min_location_failed  = 1
	restricted_roles = ["${datadog_role.bar.id}"]
}
  name    = "%[1]s"
  message = "Notify @datadog.user"
  tags    = ["multistep"]
  status  = "paused"

  config_variable {
    id   = datadog_synthetics_global_variable.global_variable.id
    type = "global"
    name = "VARIABLE_NAME"
  }

  api_step {
    name = "First api step"
    request_definition {
      method           = "GET"
      url              = "https://www.datadoghq.com"
      body             = "this is a body"
      timeout          = 30
      allow_insecure   = true
      follow_redirects = true
    }
    request_headers = {
      Accept             = "application/json"
      X-Datadog-Trace-ID = "123456789"
    }
    request_query = {
      foo = "bar"
    }
    request_basicauth {
		type = "sigv4"
		access_key = "sigv4-access-key"
		secret_key = "sigv4-secret-key"
		region = "sigv4-region"
		service_name = "sigv4-service-name"
		session_token = "sigv4-session-token"
    }
    request_client_certificate {
      cert {
        content = "content-certificate"
      }
      key {
        content  = "content-key"
        filename = "key"
      }
    }
    request_proxy {
		url = "https://proxy.url"
		headers = {
			Accept = "application/json"
			X-Datadog-Trace-ID = "123456789"
		}
	}
    assertion {
      type     = "statusCode"
      operator = "is"
      target   = "200"
    }

    extracted_value {
      name  = "VAR_EXTRACT"
      field = "content-length"
      type  = "http_header"
      parser {
        type  = "regex"
        value = ".*"
      }
      secure = true
    }
    allow_failure = true
    is_critical   = false

    retry {
      count    = 5
      interval = 1000
    }
  }
  api_step {
    name = "Second api step"
    request_definition {
      method           = "GET"
      url              = "https://docs.datadoghq.com"
      timeout          = 30
      allow_insecure   = true
      follow_redirects = true
    }
    request_basicauth {
		type = "oauth-client"
		audience = "audience"
		client_id = "client-id"
		client_secret = "client-secret"
		scope = "scope"
		token_api_authentication = "header"
		access_token_url = "https://token.datadoghq.com"
	}
    assertion {
      type     = "statusCode"
      operator = "is"
      target   = "200"
    }
  }

  api_step {
    name = "Third api step"
    request_definition {
      method           = "GET"
      url              = "https://docs.datadoghq.com"
      timeout          = 30
      allow_insecure   = true
      follow_redirects = true
    }
    request_basicauth {
		type = "oauth-rop"
		audience = "audience"
		client_id = "client-id"
		client_secret = "client-secret"
		resource = "resource"
		scope = "scope"
		token_api_authentication = "body"
		access_token_url = "https://token.datadoghq.com"
		username = "username"
		password = "password"
	}
    assertion {
      type     = "statusCode"
      operator = "is"
      target   = "200"
    }
  }

  api_step {
    name = "Fourth api step"
    request_definition {
      method           = "GET"
      url              = "https://docs.datadoghq.com"
      timeout          = 30
      allow_insecure   = true
      follow_redirects = true
    }
    request_basicauth {
		type = "digest"
		username = "username"
		password = "password"
	}
    assertion {
      type     = "statusCode"
      operator = "is"
      target   = "200"
    }
  }
}
`, testName, variableName)
}

func testSyntheticsTestExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_synthetics_test" {
				continue
			}
			if _, _, err := apiInstances.GetSyntheticsApiV1().GetTest(auth, r.Primary.ID); err != nil {
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
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if _, _, err := apiInstances.GetSyntheticsApiV1().GetTest(auth, r.Primary.ID); err != nil {
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
			apiInstances := providerConf.DatadogApiInstances
			auth := providerConf.Auth

			syntheticsTest, _, err := apiInstances.GetSyntheticsApiV1().GetBrowserTest(auth, r.Primary.ID)

			if err != nil {
				return fmt.Errorf("failed to read synthetics test %s", err)
			}

			syntheticsTestUpdate := datadogV1.NewSyntheticsBrowserTestWithDefaults()
			syntheticsTestUpdate.SetMessage(syntheticsTest.GetMessage())
			syntheticsTestUpdate.SetName(syntheticsTest.GetName())
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

			if _, _, err := apiInstances.GetSyntheticsApiV1().UpdateBrowserTest(auth, r.Primary.ID, *syntheticsTestUpdate); err != nil {
				return fmt.Errorf("failed to manually update synthetics test %s", err)
			}
		}

		return nil
	}
}
