package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogSyntheticsGlobalVariable(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToUpper(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSyntheticsGlobalVariableConfig(uniq),
				Check:  checkDatadogSyntheticsGlobalVariable(accProvider, uniq),
			},
		},
	})
}

func checkDatadogSyntheticsGlobalVariable(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
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
