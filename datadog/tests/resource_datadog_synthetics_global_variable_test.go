package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

func getUniqueVariableName(ctx context.Context, t *testing.T) string {
	return strings.ReplaceAll(strings.ToUpper(uniqueEntityName(ctx, t)), "-", "_")
}

func TestAccDatadogSyntheticsGlobalVariable_importBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	variableName := getUniqueVariableName(ctx, t)
	roleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsGlobalVariableConfig(variableName, roleName),
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableSecure_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableSecureStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariable_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableStep(ctx, accProvider, t),
			updateSyntheticsGlobalVariableStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableSecure_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableSecureStep(ctx, accProvider, t),
			updateSyntheticsGlobalVariableSecureStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableFromTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableFromTestStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableFromTest_LocalVariable(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableFromTestLocalVariableStep(ctx, accProvider, t),
		},
	})
}

func createSyntheticsGlobalVariableStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	roleName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableConfig(variableName, roleName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsResourceExists(accProvider),
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
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "restricted_roles.#", "1"),
		),
	}
}

func createSyntheticsGlobalVariableConfig(uniqVariableName string, uniqRoleName string) string {
	return fmt.Sprintf(`
resource "datadog_role" "rbac_role" {
	name = "%s"
}

resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "a global variable"
	tags = ["foo:bar", "baz"]
	value = "variable-value"
	restricted_roles = ["${datadog_role.rbac_role.id}"]
}`, uniqRoleName, uniqVariableName)
}

func updateSyntheticsGlobalVariableStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t) + "_UPDATED"
	return resource.TestStep{
		Config: updateSyntheticsGlobalVariableConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsResourceExists(accProvider),
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

func createSyntheticsGlobalVariableSecureStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableSecureConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsResourceExists(accProvider),
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
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "options.0.totp_parameters.0.digits", "6"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "options.0.totp_parameters.0.refresh_interval", "30"),
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
	options {
		totp_parameters {
			digits = 6
			refresh_interval = 30
		}
	}
}`, uniq)
}

func updateSyntheticsGlobalVariableSecureStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t) + "_UPDATED"
	return resource.TestStep{
		Config: updateSyntheticsGlobalVariableSecureConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsResourceExists(accProvider),
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
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "options.0.totp_parameters.0.digits", "6"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "options.0.totp_parameters.0.refresh_interval", "30"),
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
	options {
		totp_parameters {
			digits = 6
			refresh_interval = 30
		}
	}
}`, uniq)
}

func createSyntheticsGlobalVariableFromTestStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableFromTestConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "a global variable from http test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value", ""),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_global_variable.foo", "parse_test_id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "parse_test_options.0.type", "http_header"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "parse_test_options.0.parser.0.type", "regex"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "parse_test_options.0.parser.0.value", ".*"),
		),
	}
}

func createSyntheticsGlobalVariableFromTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "bar" {
	type = "api"
	subtype = "http"

	request_definition {
		method = "GET"
		url = "https://www.datadoghq.com"
		timeout = 30
	}

	assertion {
		type = "header"
		property = "content-type"
		operator = "contains"
		target = "application/json"
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

	name = "%[1]s"
	message = ""
	tags = []

	status = "paused"
}

resource "datadog_synthetics_global_variable" "foo" {
	name = "%[1]s"
	description = "a global variable from http test"
	tags = ["foo:bar", "baz"]
	value = ""
	parse_test_id = datadog_synthetics_test.bar.id
	parse_test_options {
		type = "http_header"
		field = "content-type"
		parser {
			type = "regex"
			value = ".*"
		}
	}
}`, uniq)
}

func createSyntheticsGlobalVariableFromTestLocalVariableStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableFromTestLocalVariableConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "a global variable from multistep test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value", ""),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_global_variable.foo", "parse_test_id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "parse_test_options.0.type", "local_variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "parse_test_options.0.local_variable_name", "LOCAL_VAR_EXTRACT"),
		),
	}
}

func createSyntheticsGlobalVariableFromTestLocalVariableConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "multi" {
	type = "api"
	subtype = "multi"

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
    assertion {
      type     = "statusCode"
      operator = "is"
      target   = "200"
    }

    extracted_value {
      name  = "LOCAL_VAR_EXTRACT"
      field = "content-length"
      type  = "http_header"
      parser {
        type  = "regex"
        value = ".*"
      }
    }
    allow_failure = true
    is_critical   = false

    retry {
      count    = 5
      interval = 1000
    }
  }

	name = "%[1]s"
	message = ""
	tags = []

	status = "paused"
}

resource "datadog_synthetics_global_variable" "foo" {
	name = "%[1]s"
	description = "a global variable from multistep test"
	tags = ["foo:bar", "baz"]
	value = ""
	parse_test_id = datadog_synthetics_test.multi.id
	parse_test_options {
		type = "local_variable"
		local_variable_name = "LOCAL_VAR_EXTRACT"
	}
}`, uniq)
}

func testSyntheticsResourceExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type == "datadog_synthetics_test" {
				if _, _, err := apiInstances.GetSyntheticsApiV1().GetTest(auth, r.Primary.ID); err != nil {
					return fmt.Errorf("received an error retrieving synthetics test %s", err)
				}
			}

			if r.Type == "datadog_synthetics_global_variable" {
				if _, _, err := apiInstances.GetSyntheticsApiV1().GetGlobalVariable(auth, r.Primary.ID); err != nil {
					return fmt.Errorf("received an error retrieving synthetics global variable %s", err)
				}
			}
		}
		return nil
	}
}

func testSyntheticsResourceIsDestroyed(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type == "datadog_role" {
				continue
			}

			if r.Type == "datadog_synthetics_test" {
				if _, _, err := apiInstances.GetSyntheticsApiV1().GetTest(auth, r.Primary.ID); err != nil {
					if strings.Contains(err.Error(), "404 Not Found") {
						continue
					}
					return fmt.Errorf("received an error retrieving synthetics test %s", err)
				}
				return fmt.Errorf("synthetics test still exists")
			}

			if _, _, err := apiInstances.GetSyntheticsApiV1().GetGlobalVariable(auth, r.Primary.ID); err != nil {
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
