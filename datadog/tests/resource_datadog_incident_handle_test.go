package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// No CheckDestroy: the API cannot delete a single handle, so destroy only removes it from state.
func TestAccDatadogIncidentHandle_Basic(t *testing.T) {
	skipIfNoCassette(t)
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := clockFromContext(ctx).Now().Unix()
	handleName := fmt.Sprintf("@incident-tf-test-%d", uniq)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentHandleConfig(uniq, handleName, "SEV-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentHandleExists(providers.frameworkProvider, "datadog_incident_handle.foo"),
					resource.TestCheckResourceAttr("datadog_incident_handle.foo", "name", handleName),
					resource.TestCheckResourceAttr("datadog_incident_handle.foo", "severity.#", "1"),
					resource.TestCheckResourceAttr("datadog_incident_handle.foo", "severity.0", "SEV-1"),
					resource.TestCheckResourceAttrPair("datadog_incident_handle.foo", "incident_type_id", "datadog_incident_type.foo", "id"),
					resource.TestCheckResourceAttrSet("datadog_incident_handle.foo", "id"),
				),
			},
			{
				Config: testAccCheckDatadogIncidentHandleConfig(uniq, handleName+"-updated", "SEV-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentHandleExists(providers.frameworkProvider, "datadog_incident_handle.foo"),
					resource.TestCheckResourceAttr("datadog_incident_handle.foo", "name", handleName+"-updated"),
					resource.TestCheckResourceAttr("datadog_incident_handle.foo", "severity.0", "SEV-2"),
				),
			},
			{
				ResourceName:      "datadog_incident_handle.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogIncidentHandleConfig(uniq int64, name string, severity string) string {
	return fmt.Sprintf(`
	resource "datadog_incident_type" "foo" {
		name        = "tf-test-handle-%d"
		description = "Incident type for incident handle tests"
	}

	resource "datadog_incident_handle" "foo" {
		name             = "%s"
		incident_type_id = datadog_incident_type.foo.id
		severity         = ["%s"]
	}`, uniq, name, severity)
}

func testAccCheckDatadogIncidentHandleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := s.RootModule().Resources[resourceName].Primary.ID
		resp, _, err := accProvider.DatadogApiInstances.GetIncidentsApiV2().ListGlobalIncidentHandles(accProvider.Auth)
		if err != nil {
			return fmt.Errorf("received an error listing incident handles: %w", err)
		}
		for _, h := range resp.Data {
			if h.GetId() == id {
				return nil
			}
		}
		return fmt.Errorf("incident handle %s not found", id)
	}
}
