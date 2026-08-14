package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogStatusPageComponent_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	prefix := "tf" + uuid.NewString()[:8]
	componentName := "datadog_status_page_component.comp"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageComponentDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageComponentConfig(uniq, prefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(componentName, "name", uniq+"-api"),
					resource.TestCheckResourceAttr(componentName, "type", "component"),
					resource.TestCheckResourceAttrSet(componentName, "group_id"),
					resource.TestCheckResourceAttrSet(componentName, "status_page_id"),
				),
			},
			{
				Config: testAccStatusPageComponentConfigRenamed(uniq, prefix),
				Check:  resource.TestCheckResourceAttr(componentName, "name", uniq+"-api-2"),
			},
			{
				ResourceName:      componentName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					r := s.RootModule().Resources[componentName]
					return fmt.Sprintf("%s:%s", r.Primary.Attributes["status_page_id"], r.Primary.ID), nil
				},
			},
		},
	})
}

func testAccStatusPageComponentConfig(uniq, prefix string) string {
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
}

resource "datadog_status_page_component" "comp" {
  status_page_id = datadog_status_page.foo.id
  name           = "%[1]s-api"
  type           = "component"
  position       = 0
  group_id       = datadog_status_page_component.grp.id
}`, uniq, prefix)
}

func testAccStatusPageComponentConfigRenamed(uniq, prefix string) string {
	return strings.Replace(testAccStatusPageComponentConfig(uniq, prefix), uniq+"-api", uniq+"-api-2", 1)
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
