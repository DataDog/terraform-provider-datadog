package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

const tfCsmThreatsAgentRulesSource = "data.datadog_csm_threats_agent_rules.acceptance_test"

func TestAccDatadogCSMThreatsAgentRulesDatasource(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCSMThreatsAgentRules(),
				Check: resource.ComposeTestCheckFunc(
					csmThreatsCheckAgentRulesCount(accProvider),
				),
			},
		},
	})
}

func csmThreatsCheckAgentRulesCount(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		agentRulesResponse, _, err := apiInstances.GetCloudWorkloadSecurityApiV2().ListCSMThreatsAgentRules(auth)
		if err != nil {
			return err
		}
		return csmThreatsAgentRulesCount(state, len(agentRulesResponse.Data))
	}
}

func csmThreatsAgentRulesCount(state *terraform.State, responseCount int) error {
	resourceAttributes := state.RootModule().Resources[tfCsmThreatsAgentRulesSource].Primary.Attributes
	agentRulesCount, _ := strconv.Atoi(resourceAttributes["agent_rules.#"])

	if agentRulesCount != responseCount {
		return fmt.Errorf("expected %d agent rules got %d agent rules",
			responseCount, agentRulesCount)
	}
	return nil
}

func testAccDataSourceCSMThreatsAgentRules() string {
	return `
data "datadog_csm_threats_agent_rules" "acceptance_test" {
}
`
}
