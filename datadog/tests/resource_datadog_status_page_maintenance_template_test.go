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

func TestAccDatadogStatusPageMaintenanceTemplate_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-maint-tmpl-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageMaintenanceTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageMaintenanceTemplateConfig(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageMaintenanceTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "name", pageName),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "maintenance_title", "Scheduled API maintenance"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "component_ids.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "scheduled_description", "Maintenance is scheduled."),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "in_progress_description", "Maintenance is in progress."),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "completed_description", "Maintenance is complete."),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_maintenance_template.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_maintenance_template.foo", "page_id"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_maintenance_template.foo", "created_at"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_maintenance_template.foo", "modified_at"),
				),
			},
		},
	})
}

func TestAccDatadogStatusPageMaintenanceTemplate_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-maint-tmpl-updated-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageMaintenanceTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageMaintenanceTemplateConfig(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageMaintenanceTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "name", pageName),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "maintenance_title", "Scheduled API maintenance"),
				),
			},
			{
				Config: testAccCheckDatadogStatusPageMaintenanceTemplateConfigUpdated(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageMaintenanceTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "name", pageName+"-updated"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "maintenance_title", "Scheduled API maintenance (updated)"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_maintenance_template.foo", "completed_description", "Maintenance has completed."),
				),
			},
		},
	})
}

func TestAccDatadogStatusPageMaintenanceTemplate_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-maint-tmpl-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageMaintenanceTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageMaintenanceTemplateConfig(pageName),
			},
			{
				ResourceName:      "datadog_status_page_maintenance_template.foo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccStatusPageMaintenanceTemplateImportStateIDFunc("datadog_status_page_maintenance_template.foo"),
			},
		},
	})
}

func testAccStatusPageMaintenanceTemplateImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found", resourceName)
		}
		return fmt.Sprintf("%s:%s", r.Primary.Attributes["page_id"], r.Primary.ID), nil
	}
}

func testAccCheckDatadogStatusPageMaintenanceTemplateConfig(pageName string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s"
  type               = "internal"
  domain_prefix      = "%[1]s"
  visualization_type = "bars_and_uptime_percentage"

  components = [
    {
      name = "API"
      type = "component"
    }
  ]
}

resource "datadog_status_page_maintenance_template" "foo" {
  page_id                  = datadog_status_page.foo.id
  name                     = "%[1]s"
  maintenance_title        = "Scheduled API maintenance"
  component_ids            = [datadog_status_page.foo.components[0].id]
  scheduled_description    = "Maintenance is scheduled."
  in_progress_description  = "Maintenance is in progress."
  completed_description    = "Maintenance is complete."
}`, pageName)
}

func testAccCheckDatadogStatusPageMaintenanceTemplateConfigUpdated(pageName string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s"
  type               = "internal"
  domain_prefix      = "%[1]s"
  visualization_type = "bars_and_uptime_percentage"

  components = [
    {
      name = "API"
      type = "component"
    }
  ]
}

resource "datadog_status_page_maintenance_template" "foo" {
  page_id                  = datadog_status_page.foo.id
  name                     = "%[1]s-updated"
  maintenance_title        = "Scheduled API maintenance (updated)"
  component_ids            = [datadog_status_page.foo.components[0].id]
  scheduled_description    = "Maintenance is scheduled."
  in_progress_description  = "Maintenance is in progress."
  completed_description    = "Maintenance has completed."
}`, pageName)
}

func testAccCheckDatadogStatusPageMaintenanceTemplateExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageMaintenanceTemplateExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogStatusPageMaintenanceTemplateDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageMaintenanceTemplateDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func statusPageMaintenanceTemplateExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page_maintenance_template" {
			continue
		}

		_, httpResp, err := api.GetMaintenanceTemplate(ctx, parseUUID(r.Primary.Attributes["page_id"]), parseUUID(r.Primary.ID))
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving maintenance template")
		}
	}
	return nil
}

func statusPageMaintenanceTemplateDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page_maintenance_template" {
			continue
		}

		_, httpResp, err := api.GetMaintenanceTemplate(ctx, parseUUID(r.Primary.Attributes["page_id"]), parseUUID(r.Primary.ID))
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving maintenance template")
		}
		return fmt.Errorf("status page maintenance template still exists")
	}
	return nil
}
