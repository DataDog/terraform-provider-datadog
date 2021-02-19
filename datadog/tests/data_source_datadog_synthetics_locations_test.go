package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDatatogSyntheticsLocation_existing(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))

	resource.ParallelTest(t, resource.TestCase{
		Providers: accProviders,
		PreCheck:  func() { testAccPreCheck(ctx, t) },
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
