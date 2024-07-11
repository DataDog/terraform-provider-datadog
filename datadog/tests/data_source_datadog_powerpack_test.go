package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogPowerpackDatasource(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourcePowerpackNameFilterConfig(uniq),
				Check:  resource.TestCheckResourceAttrSet("data.datadog_powerpack.pack_foo", "id"),
			},
		},
	})
}

func testAccPowerpackConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_powerpack" "foo" {
  description = "Created using the Datadog provider in terraform"
  name = "%s foo"
  live_span   = "4h"

  layout {
    height = 10
    width  = 3
    x      = 1
    y      = 0
  }

  template_variables {
    defaults = ["defaults"]
    name     = "datacenter"
  }

  widget {
    event_stream_definition {
      query       = "*"
      event_size  = "l"
      title       = "Widget Title"
      title_size  = 16
      title_align = "right"
    }
  }
}


resource "datadog_powerpack" "bar" {
  description = "Created using the Datadog provider in terraform"
  name = "%s bar"
  live_span   = "4h"

  layout {
    height = 10
    width  = 3
    x      = 1
    y      = 0
  }

  template_variables {
    defaults = ["defaults"]
    name     = "datacenter"
  }

  widget {
    event_stream_definition {
      query       = "*"
      event_size  = "l"
      title       = "Widget Title"
      title_size  = 16
      title_align = "right"
    }
  }
}`, uniq, uniq)
}

func testAccDatasourcePowerpackNameFilterConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_powerpack" "pack_foo" {
  depends_on = [
    datadog_powerpack.foo,
    datadog_powerpack.bar,
  ]
  name = "%s foo"
}`, testAccPowerpackConfig(uniq), uniq)
}
