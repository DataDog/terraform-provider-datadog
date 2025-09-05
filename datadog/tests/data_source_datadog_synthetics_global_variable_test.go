package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogSyntheticsGlobalVariable(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToUpper(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSyntheticsGlobalVariableConfig(uniq),
				Check:  checkDatadogSyntheticsGlobalVariable(uniq),
			},
		},
	})
}

func checkDatadogSyntheticsGlobalVariable(uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_global_variable.my_variable", "name", uniq),
		resource.TestCheckResourceAttrSet(
			"data.datadog_synthetics_global_variable.my_variable", "id"),
	)
}

func testAccDatadogSyntheticsGlobalVariableConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
  name = "%s"
  value = "bar"
}`, uniq)
}

func testAccCheckDatadogSyntheticsGlobalVariableConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_synthetics_global_variable" "my_variable" {
  depends_on = [
    datadog_synthetics_global_variable.foo,
  ]
  name = "%s"
}`, testAccDatadogSyntheticsGlobalVariableConfig(uniq), uniq)
}

func TestAccDatadogSyntheticsGlobalVariable_ParseTestOptions(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(ctx, t)

	// Generate a unique name to avoid conflicts
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, providers)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSyntheticsGlobalVariableParseTestOptions(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSyntheticsGlobalVariableExists(accProvider, "datadog_synthetics_global_variable.foo"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "value", "secret_value"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "secure", "true"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "parse_test_id", "public-abc123"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "parse_test_options.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "parse_test_options.0.field", "auth_token"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "parse_test_options.0.type", "http_header"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "parse_test_options.0.parser.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "parse_test_options.0.parser.0.type", "regex"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "parse_test_options.0.parser.0.value", "token=([a-f0-9]+)"),
				),
			},
			{
				Config: testAccCheckDatadogSyntheticsGlobalVariableParseTestOptionsUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSyntheticsGlobalVariableExists(accProvider, "datadog_synthetics_global_variable.foo"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "parse_test_options.0.field", "new_auth_token"),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_global_variable.foo", "parse_test_options.0.parser.0.value", "new_token=([a-f0-9]+)"),
				),
			},
		},
	})
}

func testAccCheckDatadogSyntheticsGlobalVariableParseTestOptions(name string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
  name = "%s"
  value = "secret_value"
  secure = true
  parse_test_id = "public-abc123"

  parse_test_options {
    field = "auth_token"
    type = "http_header"
    parser {
      type = "regex"
      value = "token=([a-f0-9]+)"
    }
  }
}
`, name)
}

func testAccCheckDatadogSyntheticsGlobalVariableParseTestOptionsUpdated(name string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "foo" {
  name = "%s"
  value = "secret_value"
  secure = true
  parse_test_id = "public-abc123"

  parse_test_options {
    field = "new_auth_token"
    type = "http_header"
    parser {
      type = "regex"
      value = "new_token=([a-f0-9]+)"
    }
  }
}
`, name)
}