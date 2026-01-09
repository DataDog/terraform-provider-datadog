package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func getUniqueVariableName(ctx context.Context, t *testing.T) string {
	return strings.ReplaceAll(strings.ToUpper(uniqueEntityName(ctx, t)), "-", "_")
}

func TestAccDatadogSyntheticsGlobalVariable_importBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	variableName := getUniqueVariableName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
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
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariable_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableStep(ctx, providers.frameworkProvider, t),
			updateSyntheticsGlobalVariableStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableSecure_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableSecureStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableSecure_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableSecureStep(ctx, providers.frameworkProvider, t),
			updateSyntheticsGlobalVariableSecureStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableTOTP_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableTOTPStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableTOTP_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableTOTPStep(ctx, providers.frameworkProvider, t),
			updateSyntheticsGlobalVariableTOTPStep(ctx, providers.frameworkProvider, t),
		},
	})
}

// fido variables
func TestAccDatadogSyntheticsGlobalVariableFIDO_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	variableName := getUniqueVariableName(ctx, t)
	config := createSyntheticsGlobalVariableConfig(variableName)
	config = strings.ReplaceAll(config, "variable-value", "fido")
	config = strings.ReplaceAll(config, "foo:bar", "fido:bar")
	config = strings.ReplaceAll(config, "baz", "fido")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableFIDOStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableFIDO_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	variableName := getUniqueVariableName(ctx, t)
	config := createSyntheticsGlobalVariableConfig(variableName)
	config = strings.ReplaceAll(config, "variable-value", "fido")
	config = strings.ReplaceAll(config, "foo:bar", "fido:bar")
	config = strings.ReplaceAll(config, "baz", "fido")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableFIDOStep(ctx, providers.frameworkProvider, t),
			updateSyntheticsGlobalVariableFIDOStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableFromTest_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableFromTestStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableFromTest_LocalVariable(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableFromTestLocalVariableStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariable_DynamicBlocks(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableDynamicBlocksStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableWriteOnly_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableWriteOnlyStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableWriteOnly_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableWriteOnlyStep(ctx, providers.frameworkProvider, t),
			updateSyntheticsGlobalVariableWriteOnlyStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsGlobalVariableWriteOnlySecure_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsGlobalVariableWriteOnlySecureStep(ctx, providers.frameworkProvider, t),
		},
	})
}

func createSyntheticsGlobalVariableStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
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

func createSyntheticsGlobalVariableConfig(uniqVariableName string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "a global variable"
	tags = ["foo:bar", "baz"]
	value = "variable-value"
}`, uniqVariableName)
}

func updateSyntheticsGlobalVariableStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t) + "_UPDATED"
	return resource.TestStep{
		Config: updateSyntheticsGlobalVariableConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
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

func createSyntheticsGlobalVariableSecureStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableSecureConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
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

func updateSyntheticsGlobalVariableSecureStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t) + "_UPDATED"
	return resource.TestStep{
		Config: updateSyntheticsGlobalVariableSecureConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
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

func createSyntheticsGlobalVariableTOTPStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableTOTPConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "a totp global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value", "variable-secure-value"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "is_totp", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "options.0.totp_parameters.0.digits", "6"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "options.0.totp_parameters.0.refresh_interval", "30"),
		),
	}
}

func createSyntheticsGlobalVariableTOTPConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "a totp global variable"
	tags = ["foo:bar", "baz"]
	value = "variable-secure-value"
	is_totp = true
	options {
		totp_parameters {
			digits = 6
			refresh_interval = 30
		}
	}
}`, uniq)
}

func updateSyntheticsGlobalVariableTOTPStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t) + "_UPDATED"
	return resource.TestStep{
		Config: updateSyntheticsGlobalVariableTOTPConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "an updated totp global variable"),
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
				"datadog_synthetics_global_variable.foo", "is_totp", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "options.0.totp_parameters.0.digits", "8"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "options.0.totp_parameters.0.refresh_interval", "60"),
		),
	}
}

func updateSyntheticsGlobalVariableTOTPConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "an updated totp global variable"
	tags = ["foo:bar", "baz", "env:test"]
	value = "variable-secure-value-updated"
	is_totp = true
	options {
		totp_parameters {
			digits = 8
			refresh_interval = 60
		}
	}
}`, uniq)
}

func createSyntheticsGlobalVariableFIDOStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableFIDOConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "a fido global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "fido:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "fido"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "is_fido", "true"),
		),
	}
}

func createSyntheticsGlobalVariableFIDOConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "a fido global variable"
	tags = ["fido:bar", "fido"]
	is_fido = true
}`, uniq)
}

func updateSyntheticsGlobalVariableFIDOStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t) + "_UPDATED"
	return resource.TestStep{
		Config: updateSyntheticsGlobalVariableFIDOConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "an updated fido global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "fido:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "fido"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "is_fido", "true"),
		),
	}
}

func updateSyntheticsGlobalVariableFIDOConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "an updated fido global variable"
	tags = ["fido:bar", "fido", "env:test"]
	is_fido = true
}`, uniq)
}

func createSyntheticsGlobalVariableFromTestStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableFromTestConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
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

func createSyntheticsGlobalVariableFromTestLocalVariableStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableFromTestLocalVariableConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
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

func testSyntheticsGlobalVariableResourceExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_synthetics_global_variable" {
				continue
			}
			if _, _, err := apiInstances.GetSyntheticsApiV1().GetGlobalVariable(auth, r.Primary.ID); err != nil {
				return fmt.Errorf("received an error retrieving synthetics global variable %s", err)
			}
		}
		return nil
	}
}

func createSyntheticsGlobalVariableDynamicBlocksStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableDynamicBlocksConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "a global variable with dynamic blocks"),
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
				"datadog_synthetics_global_variable.foo", "parse_test_options.0.field", "content-type"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "parse_test_options.0.parser.0.type", "regex"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "parse_test_options.0.parser.0.value", ".*"),
		),
	}
}

func createSyntheticsGlobalVariableDynamicBlocksConfig(uniq string) string {
	return fmt.Sprintf(`
locals {
  parse_options = [{
    type  = "http_header"
    field = "content-type"
    parser = [{
      type  = "regex"
      value = ".*"
    }]
  }]
}

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
	description = "a global variable with dynamic blocks"
	tags = ["foo:bar", "baz"]
	value = ""
	parse_test_id = datadog_synthetics_test.bar.id

	dynamic "parse_test_options" {
		for_each = local.parse_options
		content {
			type  = parse_test_options.value.type
			field = parse_test_options.value.field
			dynamic "parser" {
				for_each = parse_test_options.value.parser
				content {
					type  = parser.value.type
					value = parser.value.value
				}
			}
		}
	}
}`, uniq)
}

func testSyntheticsGlobalVariableResourceIsDestroyed(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_synthetics_global_variable" {
				continue
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

func createSyntheticsGlobalVariableWriteOnlyStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableWriteOnlyConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "a write-only global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value_wo_version", "1"),
			resource.TestCheckNoResourceAttr(
				"datadog_synthetics_global_variable.foo", "value_wo"),
		),
	}
}

func createSyntheticsGlobalVariableWriteOnlyConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "a write-only global variable"
	tags = ["foo:bar", "baz"]
	value_wo = "variable-wo-value"
	value_wo_version = "1"
}`, uniq)
}

func updateSyntheticsGlobalVariableWriteOnlyStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t) + "_UPDATED"
	return resource.TestStep{
		Config: updateSyntheticsGlobalVariableWriteOnlyConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "an updated write-only global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value_wo_version", "2"),
			resource.TestCheckNoResourceAttr(
				"datadog_synthetics_global_variable.foo", "value_wo"),
		),
	}
}

func updateSyntheticsGlobalVariableWriteOnlyConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "an updated write-only global variable"
	tags = ["foo:bar", "baz", "env:test"]
	value_wo = "variable-wo-value-updated"
	value_wo_version = "2"
}`, uniq)
}

func createSyntheticsGlobalVariableWriteOnlySecureStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	variableName := getUniqueVariableName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsGlobalVariableWriteOnlySecureConfig(variableName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsGlobalVariableResourceExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "name", variableName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "description", "a secure write-only global variable"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "secure", "true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_global_variable.foo", "value_wo_version", "1"),
			resource.TestCheckNoResourceAttr(
				"datadog_synthetics_global_variable.foo", "value_wo"),
		),
	}
}

func createSyntheticsGlobalVariableWriteOnlySecureConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
	name = "%s"
	description = "a secure write-only global variable"
	tags = ["foo:bar", "baz"]
	value_wo = "variable-wo-secure-value"
	value_wo_version = "1"
	secure = true
}`, uniq)
}
