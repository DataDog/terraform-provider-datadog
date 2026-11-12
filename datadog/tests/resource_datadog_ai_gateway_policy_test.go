package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const aiGatewayPolicyResourceName = "datadog_ai_gateway_policy.foo"

func TestAccAiGatewayPolicy(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))
	inbound := name + "-a"
	renamed := name + "-b"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAiGatewayPolicyDestroy(providers.frameworkProvider, inbound, renamed),
		Steps: []resource.TestStep{
			{
				Config: testAccAiGatewayPolicyConfig(inbound, "gpt-5.2-codex"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayPolicyExists(providers.frameworkProvider, inbound, true),
					resource.TestCheckResourceAttr(aiGatewayPolicyResourceName, "id", inbound),
					resource.TestCheckResourceAttr(aiGatewayPolicyResourceName, "priority", "0"),
					resource.TestCheckResourceAttr(aiGatewayPolicyResourceName, "tiers.#", "2"),
					resource.TestCheckResourceAttr(aiGatewayPolicyResourceName, "tiers.0.targets.0.model", "gpt-5.2-codex"),
					resource.TestCheckResourceAttr(aiGatewayPolicyResourceName, "tiers.1.targets.#", "1"),
				),
			},
			{
				// Changing a tier updates the policy in place.
				Config: testAccAiGatewayPolicyConfig(inbound, "gpt-5.1-codex"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(aiGatewayPolicyResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr(aiGatewayPolicyResourceName, "tiers.0.targets.0.model", "gpt-5.1-codex"),
			},
			{
				// Changing the inbound pattern replaces the policy, and the old one must be gone, not left live.
				Config: testAccAiGatewayPolicyConfig(renamed, "gpt-5.1-codex"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(aiGatewayPolicyResourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(aiGatewayPolicyResourceName, "id", renamed),
					testAccCheckDatadogAiGatewayPolicyExists(providers.frameworkProvider, renamed, true),
					testAccCheckDatadogAiGatewayPolicyExists(providers.frameworkProvider, inbound, false),
				),
			},
			{
				ResourceName:      aiGatewayPolicyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAiGatewayPolicyConfig(inbound, model string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_policy" "foo" {
  inbound = "%s"

  tiers = [
    {
      targets = [{ provider = "*", model = "%s" }]
    },
    {
      targets = [{ model = "claude-sonnet-5*" }]
    },
  ]
}`, inbound, model)
}

func aiGatewayPolicyExists(accProvider *fwprovider.FrameworkProvider, inbound string) (bool, error) {
	api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
	resp, _, err := api.ListAiGatewayPolicies(accProvider.Auth)
	if err != nil {
		return false, err
	}
	for _, policy := range resp.GetData() {
		if policy.GetId() == inbound {
			return true, nil
		}
	}
	return false, nil
}

func testAccCheckDatadogAiGatewayPolicyExists(accProvider *fwprovider.FrameworkProvider, inbound string, want bool) resource.TestCheckFunc {
	return func(*terraform.State) error {
		exists, err := aiGatewayPolicyExists(accProvider, inbound)
		if err != nil {
			return fmt.Errorf("received an error listing AI Gateway policies: %w", err)
		}
		if exists != want {
			return fmt.Errorf("AI Gateway policy %q exists = %t, want %t", inbound, exists, want)
		}
		return nil
	}
}

func testAccCheckDatadogAiGatewayPolicyDestroy(accProvider *fwprovider.FrameworkProvider, inbounds ...string) func(*terraform.State) error {
	return func(*terraform.State) error {
		return utils.Retry(2, 5, func() error {
			for _, inbound := range inbounds {
				exists, err := aiGatewayPolicyExists(accProvider, inbound)
				if err != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error listing AI Gateway policies: %s", err)}
				}
				if exists {
					return &utils.RetryableError{Prob: fmt.Sprintf("AI Gateway policy %q still exists", inbound)}
				}
			}
			return nil
		})
	}
}
