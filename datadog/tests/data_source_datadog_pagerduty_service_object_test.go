package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDatadogPagerdutyIntegrationSODatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationPagerdutyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourcePagerdutyIntegrationSOConfig(uniq),
				Check:  checkDatasourcePagerdutyIntegrationSOAttrs(accProvider, uniq),
			},
		},
	})
}

func checkDatasourcePagerdutyIntegrationSOAttrs(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogIntegrationPagerdutyExists(accProvider),
		resource.TestCheckResourceAttr("data.datadog_integration_pagerduty_service_object.foo", "service_name", uniq),
	)
}

func testAccPagerdutyIntegrationSOConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_pagerduty_service_object" "foo" {
	service_name = "%s"
	service_key = "%s"
}
`, uniq, uniq)
}

func testAccDatasourcePagerdutyIntegrationSOConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_integration_pagerduty_service_object" "foo" {
	depends_on = [
		datadog_integration_pagerduty_service_object.foo
	]
	service_name = "%s"
}`, testAccPagerdutyIntegrationSOConfig(uniq), uniq)
}
