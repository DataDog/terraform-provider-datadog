package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

const aiGatewayDataCaptureOverrideResourceName = "datadog_ai_gateway_data_capture_override.foo"

func TestAccAiGatewayDataCaptureOverride(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	// The override's UUID is generated client-side, so derive it from the name to replay recorded requests.
	newID := fwprovider.NewAiGatewayDataCaptureOverrideID
	fwprovider.NewAiGatewayDataCaptureOverrideID = func() uuid.UUID { return uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)) }
	t.Cleanup(func() { fwprovider.NewAiGatewayDataCaptureOverrideID = newID })

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAiGatewayDataCaptureOverrideDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// The API lowercases tag keys, so an uppercase key would show as a change on every plan.
				Config:      strings.Replace(testAccAiGatewayDataCaptureOverrideConfig(name), "ml_app =", "ML_App =", 1),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be lowercase`),
			},
			{
				Config: testAccAiGatewayDataCaptureOverrideConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayDataCaptureOverrideExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "name", name),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "mode", "metadata_only"),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "metadata_capture.#", "1"),
					resource.TestCheckTypeSetElemAttr(aiGatewayDataCaptureOverrideResourceName, "metadata_capture.*", "tool_usage"),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "tags.ml_app", strings.ToLower(name)),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "confirm_full_capture", "false"),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "revision", "1"),
				),
			},
			{
				// Turning on full capture without confirming it is rejected by the API.
				Config:      testAccAiGatewayDataCaptureOverrideConfigFullCapture(name, false),
				ExpectError: regexp.MustCompile(`confirm`),
			},
			{
				// An unset metadata_capture keeps the stored items, and a changed scope is replaced.
				Config: testAccAiGatewayDataCaptureOverrideConfigFullCapture(name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayDataCaptureOverrideExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "name", name+"-full"),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "mode", "full_capture"),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "metadata_capture.#", "1"),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "tags.provider", "openai-test"),
					resource.TestCheckResourceAttr(aiGatewayDataCaptureOverrideResourceName, "revision", "2"),
				),
			},
			{
				ResourceName:            aiGatewayDataCaptureOverrideResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"confirm_full_capture"},
			},
		},
	})
}

func testAccAiGatewayDataCaptureOverrideConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_data_capture_override" "foo" {
  name             = "%s"
  mode             = "metadata_only"
  metadata_capture = ["tool_usage"]
  tags = {
    ml_app = "%s"
  }
}`, name, strings.ToLower(name))
}

func testAccAiGatewayDataCaptureOverrideConfigFullCapture(name string, confirm bool) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_data_capture_override" "foo" {
  name                 = "%s-full"
  mode                 = "full_capture"
  confirm_full_capture = %t
  tags = {
    ml_app   = "%s"
    provider = "openai-test"
  }
}`, name, confirm, strings.ToLower(name))
}

func getAiGatewayDataCaptureOverride(accProvider *fwprovider.FrameworkProvider, id string) (*datadogV2.AiGatewayCaptureOverride, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
	resp, _, err := api.GetAiGatewayDataCaptureConfiguration(accProvider.Auth)
	if err != nil {
		return nil, err
	}
	data := resp.GetData()
	attributes := data.GetAttributes()
	for _, override := range attributes.GetOverrides() {
		if override.GetId() == parsed {
			return &override, nil
		}
	}
	return nil, nil
}

func testAccCheckDatadogAiGatewayDataCaptureOverrideExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[aiGatewayDataCaptureOverrideResourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", aiGatewayDataCaptureOverrideResourceName)
		}
		override, err := getAiGatewayDataCaptureOverride(accProvider, r.Primary.ID)
		if err != nil {
			return fmt.Errorf("received an error retrieving AI Gateway data capture override: %w", err)
		}
		if override == nil {
			return fmt.Errorf("AI Gateway data capture override %s not found", r.Primary.ID)
		}
		return nil
	}
}

func testAccCheckDatadogAiGatewayDataCaptureOverrideDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_ai_gateway_data_capture_override" {
				continue
			}
			override, err := getAiGatewayDataCaptureOverride(accProvider, r.Primary.ID)
			if err != nil {
				return fmt.Errorf("received an error retrieving AI Gateway data capture override: %w", err)
			}
			if override != nil {
				return fmt.Errorf("AI Gateway data capture override %s still exists", r.Primary.ID)
			}
		}
		return nil
	}
}
