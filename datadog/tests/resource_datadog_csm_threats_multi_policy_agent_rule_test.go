package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// Create an agent rule and update its description
func TestAccCSMThreatsMultiPolicyAgentRule_CreateAndUpdate(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	agentRuleName := uniqueAgentRuleName(ctx)
	resourceName := "datadog_csm_threats_multi_policy_agent_rule.agent_rule_test"

	policyName := uniqueAgentRuleName(ctx)
	policyConfig := fmt.Sprintf(`
	resource "datadog_csm_threats_policy" "policy_for_test" {
		name              = "%s"
		enabled           = true
		description       = "im a policy"
		tags              = ["host_name:test_host"]
	}
	`, policyName)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsMultiPolicyAgentRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Create a policy to have at least one
				Config: policyConfig,
				Check:  testAccCheckCSMThreatsPolicyExists(providers.frameworkProvider, "datadog_csm_threats_policy.policy_for_test"),
			},
			{
				Config: fmt.Sprintf(`
				%s
				resource "datadog_csm_threats_multi_policy_agent_rule" "agent_rule_test" {
					name              = "%s"
                    policy_id         = datadog_csm_threats_policy.policy_for_test.id
					enabled           = true
					description       = "im a rule"
					expression 		  = "open.file.name == \"etc/shadow/password\""
				}
				`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsMultiPolicyAgentRuleExists(providers.frameworkProvider, resourceName),
					checkCSMThreatsAgentRuleContent(
						resourceName,
						agentRuleName,
						"im a rule",
						"open.file.name == \"etc/shadow/password\"",
					),
				),
			},
			// Update description
			{
				Config: fmt.Sprintf(`
				%s
				resource "datadog_csm_threats_multi_policy_agent_rule" "agent_rule_test" {
					name              = "%s"
                    policy_id         = datadog_csm_threats_policy.policy_for_test.id
					enabled           = true
					description       = "updated agent rule for terraform provider test"
					expression 		  = "open.file.name == \"etc/shadow/password\""
				}
				`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsMultiPolicyAgentRuleExists(providers.frameworkProvider, resourceName),
					checkCSMThreatsAgentRuleContent(
						resourceName,
						agentRuleName,
						"updated agent rule for terraform provider test",
						"open.file.name == \"etc/shadow/password\"",
					),
				),
			},
		},
	})
}

func testAccCheckCSMThreatsMultiPolicyAgentRuleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in the state %s", resourceName, s.RootModule().Resources)
		}

		if resource.Type != "datadog_csm_threats_multi_policy_agent_rule" {
			return fmt.Errorf("resource %s is not of type datadog_csm_threats_multi_policy_agent_rule, found %s instead", resourceName, resource.Type)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		policyId := resource.Primary.Attributes["policy_id"]
		_, _, err := apiInstances.GetCSMThreatsApiV2().GetCSMThreatsAgentRule(auth, resource.Primary.ID, *datadogV2.NewGetCSMThreatsAgentRuleOptionalParameters().WithPolicyId(policyId))
		if err != nil {
			return fmt.Errorf("received an error retrieving agent rule: %s", err)
		}

		return nil
	}
}

func testAccCheckCSMThreatsMultiPolicyAgentRuleDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_csm_threats_multi_policy_agent_rule" {
				policyId := resource.Primary.Attributes["policy_id"]
				_, httpResponse, err := apiInstances.GetCSMThreatsApiV2().GetCSMThreatsAgentRule(auth, resource.Primary.ID, *datadogV2.NewGetCSMThreatsAgentRuleOptionalParameters().WithPolicyId(policyId))
				if err == nil {
					return errors.New("agent rule still exists")
				}
				if httpResponse == nil || httpResponse.StatusCode != 404 {
					return fmt.Errorf("received an error while getting the agent rule: %s", err)
				}
			}
		}

		return nil
	}
}
