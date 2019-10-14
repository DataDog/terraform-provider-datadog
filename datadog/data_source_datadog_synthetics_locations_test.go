package datadog

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDatatogSyntheticsLocation_existing(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				data "datadog_synthetics_locations" "test" {}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_synthetics_locations.test", ".#"),
				),
			},
		},
	})
}
