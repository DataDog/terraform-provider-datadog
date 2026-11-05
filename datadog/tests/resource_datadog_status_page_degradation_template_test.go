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

func TestAccDatadogStatusPageDegradationTemplate_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-degrade-tmpl-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageDegradationTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageDegradationTemplateConfig(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageDegradationTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "name", pageName),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "degradation_title", "API degradation"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "components_affected.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "components_affected.0.status", "degraded"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "updates.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "updates.0.status", "investigating"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_degradation_template.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_degradation_template.foo", "page_id"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_degradation_template.foo", "created_at"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_degradation_template.foo", "modified_at"),
				),
			},
		},
	})
}

func TestAccDatadogStatusPageDegradationTemplate_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-degrade-tmpl-updated-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageDegradationTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageDegradationTemplateConfig(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageDegradationTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "name", pageName),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "degradation_title", "API degradation"),
				),
			},
			{
				Config: testAccCheckDatadogStatusPageDegradationTemplateConfigUpdated(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageDegradationTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "name", pageName+"-updated"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "degradation_title", "API degradation (updated)"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_degradation_template.foo", "updates.0.status", "identified"),
				),
			},
		},
	})
}

func TestAccDatadogStatusPageDegradationTemplate_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-degrade-tmpl-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageDegradationTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageDegradationTemplateConfig(pageName),
			},
			{
				ResourceName:      "datadog_status_page_degradation_template.foo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccStatusPageDegradationTemplateImportStateIDFunc("datadog_status_page_degradation_template.foo"),
			},
		},
	})
}

func testAccStatusPageDegradationTemplateImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found", resourceName)
		}
		return fmt.Sprintf("%s:%s", r.Primary.Attributes["page_id"], r.Primary.ID), nil
	}
}

func testAccCheckDatadogStatusPageDegradationTemplateConfig(pageName string) string {
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
}

resource "datadog_status_page_degradation_template" "foo" {
  page_id           = datadog_status_page.foo.id
  name              = "%[1]s"
  degradation_title = "API degradation"

  components_affected = [
    {
      id     = datadog_status_page_component.foo.id
      status = "degraded"
    }
  ]

  updates = [
    {
      message = "We are investigating the issue."
      status   = "investigating"
    }
  ]
}`, pageName)
}

func testAccCheckDatadogStatusPageDegradationTemplateConfigUpdated(pageName string) string {
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
}

resource "datadog_status_page_degradation_template" "foo" {
  page_id           = datadog_status_page.foo.id
  name              = "%[1]s-updated"
  degradation_title = "API degradation (updated)"

  components_affected = [
    {
      id     = datadog_status_page_component.foo.id
      status = "major_outage"
    }
  ]

  updates = [
    {
      message = "We have identified the root cause."
      status   = "identified"
    }
  ]
}`, pageName)
}

func testAccCheckDatadogStatusPageDegradationTemplateExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageDegradationTemplateExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogStatusPageDegradationTemplateDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageDegradationTemplateDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func statusPageDegradationTemplateExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page_degradation_template" {
			continue
		}

		_, httpResp, err := api.GetDegradationTemplate(ctx, parseUUID(r.Primary.Attributes["page_id"]), parseUUID(r.Primary.ID))
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving degradation template")
		}
	}
	return nil
}

func statusPageDegradationTemplateDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page_degradation_template" {
			continue
		}

		_, httpResp, err := api.GetDegradationTemplate(ctx, parseUUID(r.Primary.Attributes["page_id"]), parseUUID(r.Primary.ID))
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving degradation template")
		}
		return fmt.Errorf("status page degradation template still exists")
	}
	return nil
}
