package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const tfAgentRulesSource = "data.datadog_cloud_workload_security_agent_rules.acceptance_test"

func TestAccDatadogCloudWorkloadSecurityAgentRulesDatasource(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	agentRuleName := strings.Replace(uniqueEntityName(ctx, t), "-", "_", -1)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckDatadogRoleDestroy(accProvider),
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCloudWorkloadSecurityAgentRules(agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					cloudWorkloadSecurityCheckAgentRulesCount(accProvider),
				),
			},
		},
	})
}

func cloudWorkloadSecurityCheckAgentRulesCount(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2

		agentRulesResponse, _, err := client.CloudWorkloadSecurityApi.ListCloudWorkloadSecurityAgentRules(authV2)
		if err != nil {
			return err
		}
		return cloudWorkloadSecurityAgentRulesCount(state, len(*agentRulesResponse.Data))
	}
}

func cloudWorkloadSecurityAgentRulesCount(state *terraform.State, responseCount int) error {
	resourceAttributes := state.RootModule().Resources[tfAgentRulesSource].Primary.Attributes
	agentRulesIdsCount, _ := strconv.Atoi(resourceAttributes["agent_rules_ids.#"])
	agentRulesCount, _ := strconv.Atoi(resourceAttributes["agent_rules.#"])

	if agentRulesCount != responseCount || agentRulesIdsCount != responseCount {
		return fmt.Errorf("expected %d agent rules got %d agent rules and %d agent rules ids",
			responseCount, agentRulesCount, agentRulesIdsCount)
	}
	return nil
}

func testAccDataSourceCloudWorkloadSecurityAgentRules(name string) string {
	return fmt.Sprintf(`
data "datadog_cloud_workload_security_agent_rules" "acceptance_test" {
}

resource "datadog_cloud_workload_security_agent_rule" "acceptance_test" {
    name = "%s"
    description = "an agent rule"
    enabled = false
	expression = "exec.file.name == \"go\""
}
`, name)
}
