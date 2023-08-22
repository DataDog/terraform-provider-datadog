package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogSDSGroupOrderDatasource(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDatadogSDSGroupOrderWithGroupConfig(uniq),
				// given the persistent nature of sds groups we can't check for exact value
				Check: resource.TestMatchResourceAttr("data.datadog_sensitive_data_scanner_group_order.foo", "group_ids.#", regexp.MustCompile("[^0].*$")),
			},
		},
	})
}

func testAccDatasourceDatadogSDSGroupOrderWithGroupConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_sensitive_data_scanner_group_order" "foo" {
	depends_on = [datadog_sensitive_data_scanner_group.mygroup]
}

resource "datadog_sensitive_data_scanner_group" "mygroup" {
	name        = "%s"
	description = "A relevant description"
	filter {
	  query = "service:my-service"
	}
	is_enabled   = false
	product_list = ["apm"]
}
`, uniq)
}
