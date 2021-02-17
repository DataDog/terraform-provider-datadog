package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDatatogSyntheticsLocation_existing(t *testing.T) {
	accProviders, _, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)

	resource.ParallelTest(t, resource.TestCase{
		Providers: accProviders,
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
