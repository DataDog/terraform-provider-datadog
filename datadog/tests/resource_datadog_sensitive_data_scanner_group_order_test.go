package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func GroupOrderConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_sensitive_data_scanner_group_order" "foo" {}

resource "datadog_sensitive_data_scanner_group_order" "foo" {
	groups = concat([datadog_sensitive_data_scanner_group.mygroup.id], data.datadog_sensitive_data_scanner_group_order.foo.groups)
}

resource "datadog_sensitive_data_scanner_group" "mygroup" {
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
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: GroupOrderConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"datadog_sensitive_data_scanner_group_order.foo", "groups.#"),
					resource.TestCheckResourceAttrPair(
						"datadog_sensitive_data_scanner_group_order.foo", "groups.0",
						"datadog_sensitive_data_scanner_group.mygroup", "id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
