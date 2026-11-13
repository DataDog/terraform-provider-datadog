package test

import (
	"context"
	"fmt"
	"regexp"
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

func TestAccDatadogStatusPage_TypeUpdated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-status-page-type-%d", clockFromContext(ctx).Now().Unix())
	var pageID string
	checkID := func(s *terraform.State) error {
		r, ok := s.RootModule().Resources["datadog_status_page.foo"]
		if !ok {
			return fmt.Errorf("resource datadog_status_page.foo not found")
		}
		if pageID == "" {
			pageID = r.Primary.ID
		} else if r.Primary.ID != pageID {
			return fmt.Errorf("status page ID changed from %s to %s", pageID, r.Primary.ID)
		}
		return nil
	}
	var steps []resource.TestStep
	for _, pageType := range []string{"public", "private", "public"} {
		hostPrefix := pageName
		if pageType == "private" {
			hostPrefix += ".private"
		}
		steps = append(steps, resource.TestStep{
			Config: testAccCheckDatadogStatusPageTypeConfig(pageName, pageType),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckDatadogStatusPageExists(providers.frameworkProvider),
				checkID,
				resource.TestCheckResourceAttr("datadog_status_page.foo", "type", pageType),
				resource.TestCheckResourceAttr("datadog_status_page.foo", "enabled", "false"),
				resource.TestMatchResourceAttr("datadog_status_page.foo", "page_url",
					regexp.MustCompile("^https://"+regexp.QuoteMeta(hostPrefix)+`\.statuspage\.[a-z0-9.-]+/?$`)),
			),
		})
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageDestroy(providers.frameworkProvider),
		Steps:                    steps,
	})
}

func testAccCheckDatadogStatusPageTypeConfig(pageName, pageType string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s"
  type               = "%[2]s"
  domain_prefix      = "%[1]s"
  visualization_type = "bars_and_uptime_percentage"
}`, pageName, pageType)
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
