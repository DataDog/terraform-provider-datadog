package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func GroupOrderConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_sensitive_data_scanner_group_order" "foo" {}

resource "datadog_sensitive_data_scanner_group_order" "foo" {
	group_ids = concat([datadog_sensitive_data_scanner_group.mygroup.id], data.datadog_sensitive_data_scanner_group_order.foo.group_ids)
}

resource "datadog_sensitive_data_scanner_group" "mygroup" {
	depends_on = [data.datadog_sensitive_data_scanner_group_order.foo]
	name        = "%s"
	description = "A relevant description"
	filter {
	  query = "service:my-service"
	}
	is_enabled   = false
	product_list = ["apm"]
}`, uniq)
}

func TestAccDatadogSDSGroupOrder_basic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: GroupOrderConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"datadog_sensitive_data_scanner_group_order.foo", "group_ids.#"),
					resource.TestCheckResourceAttrPair(
						"datadog_sensitive_data_scanner_group_order.foo", "group_ids.0",
						"datadog_sensitive_data_scanner_group.mygroup", "id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
