package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

const aiGatewayGovernanceResourceName = "datadog_ai_gateway_governance.foo"

// TestAccAiGatewayGovernance changes the test organization's governance controls, so it ends with the values the
// organization started with: Metadata-only capture with no metadata items. It never sets zdr_mode or full capture.
func TestAccAiGatewayGovernance(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		// Deleting the resource leaves the controls in place.
		CheckDestroy: testAccCheckDatadogAiGatewayGovernanceUnchanged(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccAiGatewayGovernanceConfig(`["tool_usage", "tool_names"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(aiGatewayGovernanceResourceName, "id", "organization"),
					resource.TestCheckResourceAttr(aiGatewayGovernanceResourceName, "data_capture_mode", "metadata_only"),
					resource.TestCheckResourceAttr(aiGatewayGovernanceResourceName, "metadata_capture.#", "2"),
					resource.TestCheckTypeSetElemAttr(aiGatewayGovernanceResourceName, "metadata_capture.*", "tool_usage"),
					resource.TestCheckTypeSetElemAttr(aiGatewayGovernanceResourceName, "metadata_capture.*", "tool_names"),
					// zdr_mode isn't configured, so it keeps the organization's value.
					resource.TestCheckResourceAttr(aiGatewayGovernanceResourceName, "zdr_mode", "off"),
				),
			},
			{
				// An empty set clears the items on the server.
				Config: testAccAiGatewayGovernanceConfig(`[]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(aiGatewayGovernanceResourceName, "metadata_capture.#", "0"),
					testAccCheckDatadogAiGatewayGovernanceMetadataCapture(providers.frameworkProvider, 0),
				),
			},
			{
				ResourceName:      aiGatewayGovernanceResourceName,
				ImportState:       true,
				ImportStateId:     "organization",
				ImportStateVerify: true,
			},
			{
				ResourceName:  aiGatewayGovernanceResourceName,
				ImportState:   true,
				ImportStateId: "not-organization",
				ExpectError:   regexp.MustCompile(`must be "organization"`),
			},
		},
	})
}

func testAccAiGatewayGovernanceConfig(metadataCapture string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_governance" "foo" {
  data_capture_mode = "metadata_only"
  metadata_capture  = %s
}`, metadataCapture)
}

func getAiGatewayGovernance(accProvider *fwprovider.FrameworkProvider) (datadogV2.AiGatewayGovernanceAttributes, error) {
	api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
	resp, _, err := api.GetAiGatewayGovernance(accProvider.Auth)
	if err != nil {
		return datadogV2.AiGatewayGovernanceAttributes{}, err
	}
	data := resp.GetData()
	return data.GetAttributes(), nil
}

func testAccCheckDatadogAiGatewayGovernanceMetadataCapture(accProvider *fwprovider.FrameworkProvider, want int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		attributes, err := getAiGatewayGovernance(accProvider)
		if err != nil {
			return fmt.Errorf("received an error retrieving AI Gateway governance: %w", err)
		}
		if got := len(attributes.GetMetadataCapture()); got != want {
			return fmt.Errorf("AI Gateway governance has %d metadata capture items, want %d", got, want)
		}
		return nil
	}
}

func testAccCheckDatadogAiGatewayGovernanceUnchanged(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(_ *terraform.State) error {
		attributes, err := getAiGatewayGovernance(accProvider)
		if err != nil {
			return fmt.Errorf("received an error retrieving AI Gateway governance: %w", err)
		}
		if mode := attributes.GetDataCaptureMode(); mode != datadogV2.AIGATEWAYDATACAPTUREMODE_METADATA_ONLY {
			return fmt.Errorf("AI Gateway governance data capture mode is %q after destroy, want it left at metadata_only", mode)
		}
		return nil
	}
}
