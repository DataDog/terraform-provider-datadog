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
