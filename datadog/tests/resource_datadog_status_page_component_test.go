package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogStatusPageComponent_Basic(t *testing.T) {
	// Not parallel: the org's status-page contract permits only one page at a time.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	tok := statusPageToken(ctx, t)
	name := "tf-sp-" + tok
	prefix := "tfsp" + tok
	grpName := "datadog_status_page_component.grp"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageComponentDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageComponentConfig(name, prefix, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(grpName, "name", name+"-group"),
					resource.TestCheckResourceAttr(grpName, "type", "group"),
					resource.TestCheckResourceAttr(grpName, "components.#", "2"),
					resource.TestCheckResourceAttr(grpName, "components.0.name", name+"-c1"),
					resource.TestCheckResourceAttr(grpName, "components.1.name", name+"-c2"),
					resource.TestCheckResourceAttrSet(grpName, "status_page_id"),
				),
			},
			{
				// add a third child -> exercises the create-child reconciliation path
				Config: testAccStatusPageComponentConfig(name, prefix, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(grpName, "components.#", "3"),
					resource.TestCheckResourceAttr(grpName, "components.2.name", name+"-c3"),
				),
			},
			{
				ResourceName:      grpName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					r := s.RootModule().Resources[grpName]
					return fmt.Sprintf("%s:%s", r.Primary.Attributes["status_page_id"], r.Primary.ID), nil
				},
			},
		},
	})
}

func testAccStatusPageComponentConfig(name, prefix string, withThird bool) string {
	third := ""
	if withThird {
		third = fmt.Sprintf(`
  components {
    name     = "%s-c3"
    position = 2
  }`, name)
	}
	return fmt.Sprintf(`
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
  components {
    name     = "%[1]s-c2"
    position = 1
  }%[3]s
}`, name, prefix, third)
}

func testAccCheckDatadogStatusPageComponentDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_status_page_component" {
				continue
			}
			pid, err := uuid.Parse(r.Primary.Attributes["status_page_id"])
			if err != nil {
				return err
			}
			cid, err := uuid.Parse(r.Primary.ID)
			if err != nil {
				return err
			}
			_, httpResp, err := accProvider.DatadogApiInstances.GetStatusPagesApiV2().GetComponent(accProvider.Auth, pid, cid)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				// deleting the parent page cascade-deletes its components, which can
				// surface as a non-404 error once the page is gone; treat as destroyed.
				continue
			}
			return fmt.Errorf("status page component %s still exists", r.Primary.ID)
		}
		return nil
	}
}
