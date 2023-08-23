package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogSyntheticsLocation_existing(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				data "datadog_synthetics_locations" "test" {}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_synthetics_locations.test", "locations.%"),
				),
			},
		},
	})
}
