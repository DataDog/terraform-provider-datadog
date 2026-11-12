package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const aiGatewayModelGroupResourceName = "datadog_ai_gateway_model_group.foo"

// TestAccAiGatewayModelGroup uses only name, allow_fast_mode, max_reasoning_level and an empty routing_rule_slugs, so
// it does not depend on model, provider or routing rule IDs that exist only in a given organization.
func TestAccAiGatewayModelGroup(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAiGatewayModelGroupDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccAiGatewayModelGroupConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayModelGroupExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayModelGroupResourceName, "name", name),
					resource.TestCheckResourceAttr(aiGatewayModelGroupResourceName, "allow_fast_mode", "false"),
					resource.TestCheckResourceAttr(aiGatewayModelGroupResourceName, "max_reasoning_level", "high"),
					resource.TestCheckResourceAttr(aiGatewayModelGroupResourceName, "routing_rule_slugs.#", "0"),
					resource.TestCheckNoResourceAttr(aiGatewayModelGroupResourceName, "model_ids"),
					resource.TestCheckNoResourceAttr(aiGatewayModelGroupResourceName, "provider_ids"),
				),
			},
			{
				Config: testAccAiGatewayModelGroupConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayModelGroupExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayModelGroupResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(aiGatewayModelGroupResourceName, "allow_fast_mode", "true"),
					resource.TestCheckNoResourceAttr(aiGatewayModelGroupResourceName, "max_reasoning_level"),
					resource.TestCheckNoResourceAttr(aiGatewayModelGroupResourceName, "routing_rule_slugs"),
					testAccCheckDatadogAiGatewayModelGroupCleared(providers.frameworkProvider),
				),
			},
			{
				ResourceName:      aiGatewayModelGroupResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAiGatewayModelGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_model_group" "foo" {
  name                = "%s"
  allow_fast_mode     = false
  max_reasoning_level = "high"
  routing_rule_slugs  = []
}`, name)
}

func testAccAiGatewayModelGroupConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_model_group" "foo" {
  name = "%s-updated"
}`, name)
}

func getAiGatewayModelGroup(accProvider *fwprovider.FrameworkProvider, id string) (datadogV2.AiGatewayModelGroupResponse, int, error) {
	api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
	resp, httpResp, err := api.GetAiGatewayModelGroup(accProvider.Auth, id)
	status := 0
	if httpResp != nil {
		status = httpResp.StatusCode
	}
	return resp, status, err
}

func testAccCheckDatadogAiGatewayModelGroupDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		return utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_ai_gateway_model_group" {
					continue
				}
				_, status, err := getAiGatewayModelGroup(accProvider, r.Primary.ID)
				if err != nil {
					if status == 404 {
						continue
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AI Gateway model group: %s", err)}
				}
				return &utils.RetryableError{Prob: "AI Gateway model group still exists"}
			}
			return nil
		})
	}
}

func testAccCheckDatadogAiGatewayModelGroupExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_ai_gateway_model_group" {
				continue
			}
			api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
			_, httpResp, err := api.GetAiGatewayModelGroup(accProvider.Auth, r.Primary.ID)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving AI Gateway model group")
			}
		}
		return nil
	}
}

// testAccCheckDatadogAiGatewayModelGroupCleared checks the API, not only state, that removing max_reasoning_level
// cleared it.
func testAccCheckDatadogAiGatewayModelGroupCleared(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_ai_gateway_model_group" {
				continue
			}
			resp, _, err := getAiGatewayModelGroup(accProvider, r.Primary.ID)
			if err != nil {
				return fmt.Errorf("error retrieving AI Gateway model group: %w", err)
			}
			attributes := resp.GetData().Attributes
			if level := attributes.MaxReasoningLevel.Get(); level != nil {
				return fmt.Errorf("max_reasoning_level is %q, want null", *level)
			}
			if slugs := attributes.GetRoutingRules(); len(slugs) > 0 {
				return fmt.Errorf("routing_rules has %d entries, want none", len(slugs))
			}
		}
		return nil
	}
}
