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

func getUniqueVariableName(clock clockwork.FakeClock, t *testing.T) string {
	return strings.ReplaceAll(strings.ToUpper(uniqueEntityName(clock, t)), "-", "_")
}

func TestAccDatadogSyntheticsGlobalVariable_importBasic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	variableName := getUniqueVariableName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsGlobalVariableIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsGlobalVariableConfig(variableName),
			},
			{
				ResourceName:      "datadog_synthetics_global_variable.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariable_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsGlobalVariableIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableSecure_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsGlobalVariableIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableSecureStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariable_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsGlobalVariableIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableStep(accProvider, clock, t),
			updateSyntheticsGlobalVariableStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableSecure_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsGlobalVariableIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableSecureStep(accProvider, clock, t),
			updateSyntheticsGlobalVariableSecureStep(accProvider, clock, t),
		},
	})
}

func createSyntheticsGlobalVariableStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "a global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value", "variable-value"),
		),
	}
}

func createSyntheticsGlobalVariableConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "a global variable"
	tags = ["foo:bar", "baz"]
	value = "variable-value"
}`, uniq)
}

func updateSyntheticsGlobalVariableStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(clock, t) + "_UPDATED"
	return resource.TestStep{
		Config: updateSyntheticsGlobalVariableConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "an updated global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value", "variable-value-updated"),
		),
	}
}

func updateSyntheticsGlobalVariableConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "an updated global variable"
	tags = ["foo:bar", "baz", "env:test"]
	value = "variable-value-updated"
}`, uniq)
}

func createSyntheticsGlobalVariableSecureStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableSecureConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "a secure global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value", "variable-secure-value"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "secure", "true"),
		),
	}
}

func createSyntheticsGlobalVariableSecureConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "a secure global variable"
	tags = ["foo:bar", "baz"]
	value = "variable-secure-value"
	secure = true
}`, uniq)
}

func updateSyntheticsGlobalVariableSecureStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(clock, t) + "_UPDATED"
	return resource.TestStep{
		Config: updateSyntheticsGlobalVariableSecureConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "an updated secure global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value", "variable-secure-value-updated"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "secure", "true"),
		),
	}
}

func updateSyntheticsGlobalVariableSecureConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "an updated secure global variable"
	tags = ["foo:bar", "baz", "env:test"]
	value = "variable-secure-value-updated"
	secure = true
}`, uniq)
}

func testSyntheticsGlobalVariableExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, _, err := datadogClientV1.SyntheticsApi.GetGlobalVariable(authV1, r.Primary.ID).Execute(); err != nil {
				return fmt.Errorf("received an error retrieving synthetics global variable %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsGlobalVariableIsDestroyed(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, _, err := datadogClientV1.SyntheticsApi.GetGlobalVariable(authV1, r.Primary.ID).Execute(); err != nil {
				if strings.Contains(err.Error(), "404 Not Found") {
					continue
				}
				return fmt.Errorf("received an error retrieving synthetics global variable %s", err)
			}
			return fmt.Errorf("synthetics global variable still exists")
		}
		return nil
	}
}
