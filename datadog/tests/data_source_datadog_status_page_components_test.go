package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogStatusPageComponentsDataSource_Basic(t *testing.T) {
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	prefix := "tf" + uuid.NewString()[:8]
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
}

resource "datadog_status_page_component" "comp" {
  status_page_id = datadog_status_page.foo.id
  name           = "%[1]s-api"
  type           = "component"
  position       = 0
  group_id       = datadog_status_page_component.grp.id
}

data "datadog_status_page_components" "all" {
  status_page_id = datadog_status_page.foo.id
  depends_on     = [datadog_status_page_component.grp, datadog_status_page_component.comp]
}`, uniq, prefix),
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
