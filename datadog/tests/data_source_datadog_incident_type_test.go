package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogIncidentTypeDataSource_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	incidentTypeName := fmt.Sprintf("test-it-ds-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentTypeDestroy(providers.frameworkProvider, "datadog_incident_type.test"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentTypeDataSourceConfig(incidentTypeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentTypeDataSourceExists(providers.frameworkProvider, "data.datadog_incident_type.foo"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_type.foo", "name", incidentTypeName),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_type.foo", "description", "Test incident type"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_type.foo", "is_default", "false"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_type.foo", "id"),
				),
			},
		},
	})
}

func testAccCheckDatadogIncidentTypeDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_incident_type" "test" {
  name        = "%s"
  description = "Test incident type"
}

data "datadog_incident_type" "foo" {
  id = datadog_incident_type.test.id
}`, name)
}

func testAccCheckDatadogIncidentTypeDataSourceExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
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
