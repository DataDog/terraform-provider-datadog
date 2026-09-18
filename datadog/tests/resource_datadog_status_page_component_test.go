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

func TestAccDatadogStatusPageComponent_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-component-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageComponentDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageComponentConfig(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageComponentExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_component.foo", "name", "API"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_component.foo", "type", "component"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_component.foo", "position", "0"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_component.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_component.foo", "page_id"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_component.foo", "status"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_component.foo", "created_at"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_component.foo", "modified_at"),
				),
			},
		},
	})
}

func TestAccDatadogStatusPageComponent_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-component-updated-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageComponentDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageComponentConfig(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageComponentExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_component.foo", "name", "API"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_component.foo", "position", "0"),
				),
			},
			{
				Config: testAccCheckDatadogStatusPageComponentConfigUpdated(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageComponentExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_component.foo", "name", "API (updated)"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_component.foo", "position", "0"),
				),
			},
		},
	})
}

func TestAccDatadogStatusPageComponent_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-component-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageComponentDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageComponentConfig(pageName),
			},
			{
				ResourceName:      "datadog_status_page_component.foo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccStatusPageComponentImportStateIDFunc("datadog_status_page_component.foo"),
			},
		},
	})
}

func testAccStatusPageComponentImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found", resourceName)
		}
		return fmt.Sprintf("%s:%s", r.Primary.Attributes["page_id"], r.Primary.ID), nil
	}
}

func testAccCheckDatadogStatusPageComponentConfig(pageName string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s"
  type               = "internal"
  domain_prefix      = "%[1]s"
  visualization_type = "bars_and_uptime_percentage"
}

resource "datadog_status_page_component" "foo" {
  page_id  = datadog_status_page.foo.id
  name     = "API"
  type     = "component"
  position = 0
}`, pageName)
}

func testAccCheckDatadogStatusPageComponentConfigUpdated(pageName string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s"
  type               = "internal"
  domain_prefix      = "%[1]s"
  visualization_type = "bars_and_uptime_percentage"
}

resource "datadog_status_page_component" "foo" {
  page_id  = datadog_status_page.foo.id
  name     = "API (updated)"
  type     = "component"
  position = 0
}`, pageName)
}

func testAccCheckDatadogStatusPageComponentExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageComponentExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogStatusPageComponentDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageComponentDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func statusPageComponentExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page_component" {
			continue
		}

		_, httpResp, err := api.GetComponent(ctx, parseUUID(r.Primary.Attributes["page_id"]), parseUUID(r.Primary.ID))
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving status page component")
		}
	}
	return nil
}

func statusPageComponentDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page_component" {
			continue
		}

		_, httpResp, err := api.GetComponent(ctx, parseUUID(r.Primary.Attributes["page_id"]), parseUUID(r.Primary.ID))
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving status page component")
		}
		return fmt.Errorf("status page component still exists")
	}
	return nil
}
