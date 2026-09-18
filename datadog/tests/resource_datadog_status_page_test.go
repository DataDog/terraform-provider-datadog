package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogStatusPage_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-status-page-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageConfig(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page.foo", "name", pageName),
					resource.TestCheckResourceAttr(
						"datadog_status_page.foo", "type", "internal"),
					resource.TestCheckResourceAttr(
						"datadog_status_page.foo", "visualization_type", "bars_and_uptime_percentage"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page.foo", "domain_prefix"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page.foo", "page_url"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page.foo", "created_at"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page.foo", "modified_at"),
				),
			},
		},
	})
}

func TestAccDatadogStatusPage_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-status-page-updated-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageConfig(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page.foo", "name", pageName),
					resource.TestCheckResourceAttr(
						"datadog_status_page.foo", "visualization_type", "bars_and_uptime_percentage"),
				),
			},
			{
				Config: testAccCheckDatadogStatusPageConfigUpdated(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page.foo", "name", pageName+"-updated"),
					resource.TestCheckResourceAttr(
						"datadog_status_page.foo", "visualization_type", "bars_only"),
				),
			},
		},
	})
}

func TestAccDatadogStatusPage_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-status-page-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageConfig(pageName),
			},
			{
				ResourceName:      "datadog_status_page.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogStatusPageConfig(pageName string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s"
  type               = "internal"
  domain_prefix      = "%[1]s"
  visualization_type = "bars_and_uptime_percentage"
}`, pageName)
}

func testAccCheckDatadogStatusPageConfigUpdated(pageName string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s-updated"
  type               = "internal"
  domain_prefix      = "%[1]s"
  visualization_type = "bars_only"
}`, pageName)
}

func testAccCheckDatadogStatusPageExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogStatusPageDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func statusPageExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := api.GetStatusPage(ctx, parseUUID(id))
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving status page")
		}
	}
	return nil
}

func statusPageDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := api.GetStatusPage(ctx, parseUUID(id))
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving status page")
		}
		return fmt.Errorf("status page still exists")
	}
	return nil
}
