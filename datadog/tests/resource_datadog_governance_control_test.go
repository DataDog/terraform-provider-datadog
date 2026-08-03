package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogGovernanceControl_Basic(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogGovernanceControlConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGovernanceControlExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "detection_type", "unused_api_keys"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "detection_frequency", "daily"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "id", "unused_api_keys"),
					resource.TestCheckResourceAttrSet(
						"datadog_governance_control.foo", "name"),
				),
			},
			{
				Config: testAccCheckDatadogGovernanceControlConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGovernanceControlExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "detection_frequency", "daily"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "name", "Unused API Keys"),
				),
			},
			{
				ResourceName:      "datadog_governance_control.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogGovernanceControlConfig() string {
	return `
resource "datadog_governance_control" "foo" {
  detection_type      = "unused_api_keys"
  detection_frequency = "daily"
}`
}

func testAccCheckDatadogGovernanceControlConfigUpdated() string {
	return `
resource "datadog_governance_control" "foo" {
  detection_type      = "unused_api_keys"
  detection_frequency = "daily"
  name                = "Unused API Keys"
}`
}

func testAccCheckDatadogGovernanceControlExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_governance_control" {
				continue
			}
			detectionType := r.Primary.ID
			_, httpResp, err := apiInstances.GetGovernanceControlsApiV2().GetGovernanceControl(auth, detectionType)
			if err != nil {
				return fmt.Errorf("received an error retrieving governance control %s: %s", detectionType, err)
			}
			if httpResp.StatusCode != 200 {
				return fmt.Errorf("governance control %s not found", detectionType)
			}
		}
		return nil
	}
}
