package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogIncidentType_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	incidentTypeName := fmt.Sprintf("test-it-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentTypeDestroy(providers.frameworkProvider, "datadog_incident_type.foo"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentTypeConfig(incidentTypeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentTypeExists(providers.frameworkProvider, "datadog_incident_type.foo"),
					resource.TestCheckResourceAttr(
						"datadog_incident_type.foo", "name", incidentTypeName),
					resource.TestCheckResourceAttr(
						"datadog_incident_type.foo", "description", "Test incident type"),
					resource.TestCheckResourceAttr(
						"datadog_incident_type.foo", "is_default", "false"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_type.foo", "id"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentType_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	incidentTypeName := fmt.Sprintf("test-it-updated-%d", clockFromContext(ctx).Now().Unix())
	incidentTypeNameUpdated := incidentTypeName + "-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentTypeDestroy(providers.frameworkProvider, "datadog_incident_type.foo"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentTypeConfig(incidentTypeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentTypeExists(providers.frameworkProvider, "datadog_incident_type.foo"),
					resource.TestCheckResourceAttr(
						"datadog_incident_type.foo", "name", incidentTypeName),
					resource.TestCheckResourceAttr(
						"datadog_incident_type.foo", "description", "Test incident type"),
				),
			},
			{
				Config: testAccCheckDatadogIncidentTypeConfigUpdated(incidentTypeNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentTypeExists(providers.frameworkProvider, "datadog_incident_type.foo"),
					resource.TestCheckResourceAttr(
						"datadog_incident_type.foo", "name", incidentTypeNameUpdated),
					resource.TestCheckResourceAttr(
						"datadog_incident_type.foo", "description", "Updated test incident type"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentType_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	incidentTypeName := fmt.Sprintf("test-it-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentTypeDestroy(providers.frameworkProvider, "datadog_incident_type.foo"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentTypeConfig(incidentTypeName),
			},
			{
				ResourceName:      "datadog_incident_type.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogIncidentTypeConfig(name string) string {
	return fmt.Sprintf(`
	resource "datadog_incident_type" "foo" {
		name = "%s"
		description = "Test incident type"
	}`, name)
}

func testAccCheckDatadogIncidentTypeConfigUpdated(name string) string {
	return fmt.Sprintf(`
	resource "datadog_incident_type" "foo" {
		name        = "%s"
		description = "Updated test incident type"
		is_default  = false
	}`, name)
}

func testAccCheckDatadogIncidentTypeExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		id := s.RootModule().Resources[resourceName].Primary.ID
		if _, _, err := apiInstances.GetIncidentsApiV2().GetIncidentType(auth, id); err != nil {
			return fmt.Errorf("received an error retrieving incident type %s", err)
		}
		return nil
	}
}

func testAccCheckDatadogIncidentTypeDestroy(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resource := s.RootModule().Resources[resourceName]
		_, httpResp, err := apiInstances.GetIncidentsApiV2().GetIncidentType(auth, resource.Primary.ID)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return nil
			}
			return err
		}

		return nil
	}
}
