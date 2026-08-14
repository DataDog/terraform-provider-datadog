package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogStatusPageComponentsDataSource_Basic(t *testing.T) {
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	tok := statusPageToken(ctx, t)
	name := "tf-sp-" + tok
	prefix := "tfsp" + tok
	dsName := "data.datadog_status_page_components.all"

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

resource "datadog_status_page_component" "grp" {
  status_page_id = datadog_status_page.foo.id
  name           = "%[1]s-group"
  type           = "group"
  position       = 0

  components {
    name     = "%[1]s-c1"
    position = 0
  }
}

data "datadog_status_page_components" "all" {
  status_page_id = datadog_status_page.foo.id
  depends_on     = [datadog_status_page_component.grp]
}`, name, prefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dsName, "components.#", "2"),
					resource.TestCheckResourceAttr(dsName, "components.0.type", "group"),
					resource.TestCheckResourceAttr(dsName, "components.1.type", "component"),
					resource.TestCheckResourceAttrSet(dsName, "components.1.group_id"),
				),
			},
		},
	})
}
