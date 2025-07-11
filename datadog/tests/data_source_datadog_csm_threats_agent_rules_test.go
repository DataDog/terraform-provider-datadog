package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccCSMThreatsAgentRulesDataSource(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	policyName := uniqueAgentRuleName(ctx)
	policyConfig := fmt.Sprintf(`
	resource "datadog_csm_threats_policy" "policy_for_test" {
		name              = "%s"
		enabled           = true
		description       = "im a policy"
		tags              = ["host_name:test_host"]
	}
	`, policyName)

	agentRuleName := uniqueAgentRuleName(ctx)
	agentRuleConfig := fmt.Sprintf(`
		%s
		resource "datadog_csm_threats_agent_rule" "agent_rule_for_data_source_test" {
			name              = "%s"
			policy_id         = datadog_csm_threats_policy.policy_for_test.id
			enabled           = true
			description       = "im a rule"
			expression 		  = "open.file.name == \"etc/shadow/password\""
			product_tags      = ["compliance_framework:PCI-DSS"]
			actions {
				set {
					name   = "test_action"
					field  = "exec.file.path"
					append = false
					scope  = "process"
				}
				hash {}
			}
		}
	`, policyConfig, agentRuleName)
	dataSourceName := "data.datadog_csm_threats_agent_rules.my_data_source"

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
				data "datadog_csm_threats_agent_rules" "my_data_source" {
					policy_id = datadog_csm_threats_policy.policy_for_test.id
				}
				`, agentRuleConfig),
				Check: checkCSMThreatsAgentRulesDataSourceContent(providers.frameworkProvider, dataSourceName, agentRuleName),
			},
		},
	})
}

func testAccCheckDatadogCSMThreatsAgentRulesDataSourceConfig(policyName, agentRuleName string) string {
	return fmt.Sprintf(`
data "datadog_csm_threats_agent_rules" "my_data_source" {
  policy_id = datadog_csm_threats_policy.policy.id
}

resource "datadog_csm_threats_policy" "policy" {
  name = "%s"
  enabled = true
  description = "Test description"
  tags = ["host_name:test_host"]
}

resource "datadog_csm_threats_agent_rule" "agent_rule" {
  name = "%s"
  description = "Test description"
  enabled = true
  expression = "open.file.name == \"etc/shadow/password\""
  policy_id = datadog_csm_threats_policy.policy.id
  product_tags = ["compliance_framework:PCI-DSS"]
  actions {
    set {
      name   = "test_action"
      field  = "exec.file.path"
      append = false
      scope  = "process"
    }
    hash {}
  }
}
`, policyName, agentRuleName)
}

func checkCSMThreatsAgentRulesDataSourceContent(accProvider *fwprovider.FrameworkProvider, dataSourceName string, agentRuleName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("resource missing from state: %s", dataSourceName)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		policyId := res.Primary.Attributes["policy_id"]
		allAgentRulesResponse, _, err := apiInstances.GetCSMThreatsApiV2().ListCSMThreatsAgentRules(auth, *datadogV2.NewListCSMThreatsAgentRulesOptionalParameters().WithPolicyId(policyId))
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

		// Check that the data_source fetched is correct
		resourceAttributes := res.Primary.Attributes
		agentRulesIdsCount, err := strconv.Atoi(resourceAttributes["agent_rules_ids.#"])
		if err != nil {
			return err
		}
		agentRulesCount, err := strconv.Atoi(resourceAttributes["agent_rules.#"])
		if err != nil {
			return err
		}
		if agentRulesCount != agentRulesIdsCount {
			return fmt.Errorf("the data source contains %d agent rules IDs but %d agent rules", agentRulesIdsCount, agentRulesCount)
		}

		// Find in which position is the agent rule we created, and check its values
		idx := 0
		for idx < agentRulesIdsCount && resourceAttributes[fmt.Sprintf("agent_rules_ids.%d", idx)] != agentRuleId {
			idx++
		}
		if idx == len(resourceAttributes) {
			return fmt.Errorf("agent rule with ID '%s' not found in data source", agentRuleId)
		}

		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.name", idx), ruleName),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.enabled", idx), "true"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.description", idx), "im a rule"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.expression", idx), "open.file.name == \"etc/shadow/password\""),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.product_tags.#", idx), "1"),
			resource.TestCheckTypeSetElemAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.product_tags.*", idx), "compliance_framework:PCI-DSS"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.actions.#", idx), "1"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.actions.0.set.name", idx), "test_action"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.actions.0.set.field", idx), "exec.file.path"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.actions.0.set.append", idx), "false"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.actions.0.set.scope", idx), "process"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("agent_rules.%d.actions.0.hash.%%", idx), "0"),
		)(state)
	}
}
