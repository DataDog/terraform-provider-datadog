package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDatatogSyntheticsLocation_existing(t *testing.T) {
	t.Parallel()
	ctx := testSpan(context.Background(), t)
	ctx, accProviders := testAccProviders(ctx, t, initRecorder(t))

	resource.Test(t, resource.TestCase{
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
