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
	skipIfNoCassette(t)
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Basic: only detection_type set, everything else computed from the control's
				// existing configuration.
				Config: testAccCheckDatadogGovernanceControlConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGovernanceControlExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "detection_type", "unused_api_keys"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "id", "unused_api_keys"),
					resource.TestCheckResourceAttrSet(
						"datadog_governance_control.foo", "name"),
				),
			},
			{
				// Explicitly manage detection_parameters, mitigation_type, and mitigation_parameters.
				Config: testAccCheckDatadogGovernanceControlConfigWithParameters(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGovernanceControlExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "mitigation_type", "revoke_api_key"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "mitigation_parameters", "{}"),
					testAccCheckDatadogGovernanceControlDetectionParameter(
						providers.frameworkProvider, "api_key_threshold", float64(45)),
					testAccCheckDatadogGovernanceControlDetectionParameter(
						providers.frameworkProvider, "governance_remediation_delay", float64(14)),
				),
			},
			{
				// Add notification_settings with multiple targets for a single event type.
				Config: testAccCheckDatadogGovernanceControlConfigWithNotificationSettings(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGovernanceControlExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "notification_settings.0.event_type", "new_detection"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "notification_settings.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "notification_settings.0.targets.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "notification_settings.0.targets.0.type", "email"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "notification_settings.0.targets.0.handle", "test@datadoghq.com"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "notification_settings.0.targets.1.type", "email"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "notification_settings.0.targets.1.handle", "test2@datadoghq.com"),
				),
			},
			{
				// Remove notification_settings from config: since it's Optional+Computed, the
				// prior value should stick around unchanged rather than reverting to empty.
				Config: testAccCheckDatadogGovernanceControlConfigWithParameters(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGovernanceControlExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "notification_settings.0.event_type", "new_detection"),
					resource.TestCheckResourceAttr(
						"datadog_governance_control.foo", "notification_settings.0.targets.#", "2"),
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
  detection_type = "unused_api_keys"
}`
}

func testAccCheckDatadogGovernanceControlConfigWithParameters() string {
	return `
resource "datadog_governance_control" "foo" {
  detection_type = "unused_api_keys"

  detection_parameters = jsonencode({
    api_key_threshold            = 45
    governance_remediation_delay = 14
  })

  mitigation_type       = "revoke_api_key"
  mitigation_parameters = jsonencode({})
}`
}

func testAccCheckDatadogGovernanceControlConfigWithNotificationSettings() string {
	return `
resource "datadog_governance_control" "foo" {
  detection_type = "unused_api_keys"

  detection_parameters = jsonencode({
    api_key_threshold            = 45
    governance_remediation_delay = 14
  })

  mitigation_type       = "revoke_api_key"
  mitigation_parameters = jsonencode({})

  notification_settings = [
    {
      event_type = "new_detection"
      enabled    = true
      targets = [
        {
          type   = "email"
          handle = "test@datadoghq.com"
        },
        {
          type   = "email"
          handle = "test2@datadoghq.com"
        }
      ]
    }
  ]
}`
}

// testAccCheckDatadogGovernanceControlDetectionParameter asserts that a detection parameter has
// the given value in the control's actual (API-side) detection_parameters, independent of what
// Terraform state reports.
func testAccCheckDatadogGovernanceControlDetectionParameter(accProvider *fwprovider.FrameworkProvider, name string, expected interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resp, _, err := apiInstances.GetGovernanceConsoleApiV2().GetGovernanceControl(auth, "unused_api_keys")
		if err != nil {
			return fmt.Errorf("error retrieving governance control: %w", err)
		}
		attributes := resp.Data.GetAttributes()
		value, ok := attributes.GetDetectionParameters()[name]
		if !ok {
			return fmt.Errorf("detection parameter %q not set", name)
		}
		if value != expected {
			return fmt.Errorf("detection parameter %q: expected %v, got %v", name, expected, value)
		}
		return nil
	}
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
			_, httpResp, err := apiInstances.GetGovernanceConsoleApiV2().GetGovernanceControl(auth, detectionType)
			if err != nil {
				return fmt.Errorf("received an error retrieving governance control %s: %w", detectionType, err)
			}
			if httpResp.StatusCode != 200 {
				return fmt.Errorf("governance control %s not found", detectionType)
			}
		}
		return nil
	}
}
