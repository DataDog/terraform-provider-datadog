package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccCSMThreatsAgentRuleDataSource(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	agentRuleName := uniqueEntityName(ctx, t)
	dataSourceName := "data.datadog_csm_threats_agent_rule.my_data_source"

	agentRuleConfig := fmt.Sprintf(`
	resource "datadog_csm_threats_agent_rule" "agent_rule_for_data_source_test" {
		name              = "%s"
		enabled           = true
		description       = "im a rule"
		expression 		  = "open.file.name == \"etc/shadow/password\""
	}
	`, agentRuleName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsAgentRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Create an agent rule to have at least one
				Config: agentRuleConfig,
				Check:  testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, "datadog_csm_threats_agent_rule.agent_rule_for_data_source_test"),
			},
			{
				Config: fmt.Sprintf(`
				%s
				data "datadog_csm_threats_agent_rules" "my_data_source" {}
				`, agentRuleConfig),
				Check: checkCSMThreatsAgentRulesDataSourceContent(providers.frameworkProvider, dataSourceName, agentRuleName),
			},
		},
	})
}

func checkCSMThreatsAgentRulesDataSourceContent(accProvider *fwprovider.FrameworkProvider, dataSourceName string, agentRuleName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("resource missing from state: %s", dataSourceName)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		allAgentRulesResponse, _, err := apiInstances.GetCloudWorkloadSecurityApiV2().ListCSMThreatsAgentRules(auth)
		if err != nil {
			return err
		}

		// Check the agentRule we created is in the API response
		agentRuleId := ""
		ruleName := ""
		for _, rule := range allAgentRulesResponse.GetData() {
			if rule.Attributes.GetName() == agentRuleName {
				agentRuleId = rule.GetId()
				ruleName = rule.Attributes.GetName()
				break
			}
		}

		if agentRuleId == "" {
			return fmt.Errorf("agent rule with name '%s' not found in API responses", agentRuleName)
		}

		resourceAttributes := res.Primary.Attributes
		idx := 0
		for idx < len(resourceAttributes) && resourceAttributes[fmt.Sprintf("agent_rules.%d", idx)] != agentRuleId {
			idx++
		}

		if idx == len(resourceAttributes) {
			return fmt.Errorf("agent rule with ID '%s' not found in data source", agentRuleId)
		}

		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(dataSourceName, ruleName, agentRuleName),
		)(state)
	}
}