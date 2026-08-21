package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogStatusPageDataSource_Basic(t *testing.T) {
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	tok := statusPageToken(ctx, t)
	name := "tf-sp-" + tok
	prefix := "tfsp" + tok

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s"
  domain_prefix      = "%[2]s"
  type               = "internal"
  visualization_type = "bars_only"
}

data "datadog_status_page" "by_id" {
  id = datadog_status_page.foo.id
}`, name, prefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_status_page.by_id", "name", name),
					resource.TestCheckResourceAttr("data.datadog_status_page.by_id", "type", "internal"),
					resource.TestCheckResourceAttr("data.datadog_status_page.by_id", "visualization_type", "bars_only"),
				),
			},
		},
	})
}
