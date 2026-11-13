package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const aiGatewayRoutingRuleResourceName = "datadog_ai_gateway_routing_rule.foo"

// aiGatewayRoutingRuleTestProvider returns the provider and two models the test routes to. Model targets must name a model
// enabled on a provider in the test org, and a provider needs an integration account created in the Datadog console, so
// recording the cassette requires DD_TEST_AI_GATEWAY_PROVIDER, DD_TEST_AI_GATEWAY_MODEL and DD_TEST_AI_GATEWAY_MODEL_2 to
// name them. The defaults are the provider and models the cassette was recorded with. CI's live runs use the same
// openai-test provider in the test org, so it must not be deleted.
func aiGatewayRoutingRuleTestProvider() (provider, model, model2 string) {
	provider, model, model2 = "openai-test", "gpt-4o-mini", "gpt-4o"
	if v := os.Getenv("DD_TEST_AI_GATEWAY_PROVIDER"); v != "" {
		provider = v
	}
	if v := os.Getenv("DD_TEST_AI_GATEWAY_MODEL"); v != "" {
		model = v
	}
	if v := os.Getenv("DD_TEST_AI_GATEWAY_MODEL_2"); v != "" {
		model2 = v
	}
	return provider, model, model2
}

func TestAccAiGatewayRoutingRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	slug := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))
	provider, model, model2 := aiGatewayRoutingRuleTestProvider()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAiGatewayRoutingRuleDestroy(providers.frameworkProvider, slug),
		Steps: []resource.TestStep{
			{
				Config: testAccAiGatewayRoutingRuleWeightedConfig(slug, provider, model, model2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayRoutingRuleTargets(providers.frameworkProvider, slug, 2),
					resource.TestCheckResourceAttr(aiGatewayRoutingRuleResourceName, "id", slug),
					resource.TestCheckResourceAttrSet(aiGatewayRoutingRuleResourceName, "route_id"),
					resource.TestCheckResourceAttr(aiGatewayRoutingRuleResourceName, "tiers.0.ordering", "weighted_affinity"),
					resource.TestCheckResourceAttr(aiGatewayRoutingRuleResourceName, "tiers.0.targets.0.weight", "70"),
					resource.TestCheckNoResourceAttr(aiGatewayRoutingRuleResourceName, "tiers.0.targets.1.weight"),
				),
			},
			{
				// Removing a target updates the rule in place, and the server drops the target.
				Config: testAccAiGatewayRoutingRuleConfiguredConfig(slug, provider, model),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(aiGatewayRoutingRuleResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayRoutingRuleTargets(providers.frameworkProvider, slug, 1),
					resource.TestCheckResourceAttr(aiGatewayRoutingRuleResourceName, "tiers.0.ordering", "configured"),
					resource.TestCheckResourceAttr(aiGatewayRoutingRuleResourceName, "tiers.0.targets.#", "1"),
				),
			},
			{
				ResourceName:      aiGatewayRoutingRuleResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAiGatewayRoutingRuleWeightedConfig(slug, provider, model, model2 string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_routing_rule" "foo" {
  slug = "%s"

  tiers = [
    {
      ordering = "weighted_affinity"
      targets = [
        { provider = "%s", model = "%s", weight = 70 },
        { provider = "%s", model = "%s" },
      ]
    },
  ]
}`, slug, provider, model, provider, model2)
}

func testAccAiGatewayRoutingRuleConfiguredConfig(slug, provider, model string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_routing_rule" "foo" {
  slug = "%s"

  tiers = [
    {
      targets = [{ provider = "%s", model = "%s" }]
    },
  ]
}`, slug, provider, model)
}

// testAccCheckDatadogAiGatewayRoutingRuleTargets checks the number of targets the server holds, so a target removed from
// the configuration must be gone from the API, not only from state.
func testAccCheckDatadogAiGatewayRoutingRuleTargets(accProvider *fwprovider.FrameworkProvider, slug string, want int) resource.TestCheckFunc {
	return func(*terraform.State) error {
		api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
		resp, _, err := api.GetAiGatewayRoutingRule(accProvider.Auth, slug)
		if err != nil {
			return fmt.Errorf("received an error retrieving AI Gateway routing rule: %w", err)
		}
		got := 0
		for _, tier := range resp.Data.Attributes.GetTiers() {
			got += len(tier.GetTargets())
		}
		if got != want {
			return fmt.Errorf("AI Gateway routing rule %q has %d targets, want %d", slug, got, want)
		}
		return nil
	}
}

func testAccCheckDatadogAiGatewayRoutingRuleDestroy(accProvider *fwprovider.FrameworkProvider, slug string) func(*terraform.State) error {
	return func(*terraform.State) error {
		api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
		return utils.Retry(2, 5, func() error {
			_, httpResp, err := api.GetAiGatewayRoutingRule(accProvider.Auth, slug)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AI Gateway routing rule: %s", err)}
			}
			return &utils.RetryableError{Prob: fmt.Sprintf("AI Gateway routing rule %q still exists", slug)}
		})
	}
}
