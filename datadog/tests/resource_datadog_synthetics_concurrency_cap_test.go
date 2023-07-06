package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogSyntheticsConcurrencyCap(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "datadog_synthetics_concurrency_cap" "this" {
					on_demand_concurrency_cap = 2
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_synthetics_concurrency_cap.this", "on_demand_concurrency_cap", "2"),
				),
			},
		},
	})
}
