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

func TestAccDatadogStatusPage_Basic(t *testing.T) {
	// Not parallel: the org's status-page contract permits only one page at a time.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	prefix := "tf" + uuid.NewString()[:8]
	resourceName := "datadog_status_page.foo"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccStatusPageConfig(uniq, prefix, "internal", "bars_and_uptime_percentage"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "type", "internal"),
					resource.TestCheckResourceAttr(resourceName, "visualization_type", "bars_and_uptime_percentage"),
					resource.TestCheckResourceAttrSet(resourceName, "page_url"),
				),
			},
			{
				// mutate patchable fields -> must NOT force replacement
				Config: testAccStatusPageConfig(uniq+"-2", prefix, "internal", "bars_only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq+"-2"),
					resource.TestCheckResourceAttr(resourceName, "visualization_type", "bars_only"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccStatusPageConfig(name, prefix, pageType, viz string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%s"
  domain_prefix      = "%s"
  type               = "%s"
  visualization_type = "%s"
}`, name, prefix, pageType, viz)
}

func testAccCheckDatadogStatusPageExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := s.RootModule().Resources[n].Primary.ID
		pid, err := uuid.Parse(id)
		if err != nil {
			return err
		}
		if _, _, err := accProvider.DatadogApiInstances.GetStatusPagesApiV2().GetStatusPage(accProvider.Auth, pid); err != nil {
			return fmt.Errorf("error retrieving status page %s: %w", id, err)
		}
		return nil
	}
}

func testAccCheckDatadogStatusPageDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_status_page" {
				continue
			}
			pid, err := uuid.Parse(r.Primary.ID)
			if err != nil {
				return err
			}
			_, httpResp, err := accProvider.DatadogApiInstances.GetStatusPagesApiV2().GetStatusPage(accProvider.Auth, pid)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("error retrieving status page: %w", err)
			}
			return fmt.Errorf("status page %s still exists", r.Primary.ID)
		}
		return nil
	}
}
