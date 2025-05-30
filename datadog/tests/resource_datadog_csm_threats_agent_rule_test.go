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
func TestAccCSMThreatsAgentRule_CreateAndUpdate(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	agentRuleName := uniqueAgentRuleName(ctx)
	resourceName := "datadog_csm_threats_agent_rule.agent_rule_test"

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
		CheckDestroy:             testAccCheckCSMThreatsAgentRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Create a policy to have at least one
				Config: policyConfig,
				Check:  testAccCheckCSMThreatsPolicyExists(providers.frameworkProvider, "datadog_csm_threats_policy.policy_for_test"),
			},
			{
				Config: fmt.Sprintf(`
				%s
				resource "datadog_csm_threats_agent_rule" "agent_rule_test" {
					name              = "%s"
                    policy_id         = datadog_csm_threats_policy.policy_for_test.id
					enabled           = true
					description       = "im a rule"
					expression 		  = "open.file.name == \"etc/shadow/password\""
					product_tags      = ["compliance_framework:PCI-DSS"]
				}
				`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, resourceName),
					checkCSMThreatsAgentRuleContent(
						resourceName,
						agentRuleName,
						"im a rule",
						"open.file.name == \"etc/shadow/password\"",
						"compliance_framework:PCI-DSS",
					),
				),
			},
			// Update description
			{
				Config: fmt.Sprintf(`
				%s
				resource "datadog_csm_threats_agent_rule" "agent_rule_test" {
					name              = "%s"
                    policy_id         = datadog_csm_threats_policy.policy_for_test.id
					enabled           = true
					description       = "updated agent rule for terraform provider test"
					expression 		  = "open.file.name == \"etc/shadow/password\""
					product_tags      = ["compliance_framework:ISO-27799"]
				}
				`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, resourceName),
					checkCSMThreatsAgentRuleContent(
						resourceName,
						agentRuleName,
						"updated agent rule for terraform provider test",
						"open.file.name == \"etc/shadow/password\"",
						"compliance_framework:ISO-27799",
					),
				),
			},
		},
	})
}

func TestAccCSMThreatsAgentRule_CreateAndUpdateWithoutPolicyID(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	agentRuleName := uniqueAgentRuleName(ctx)
	resourceName := "datadog_csm_threats_agent_rule.agent_rule_without_policy"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsAgentRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "datadog_csm_threats_agent_rule" "agent_rule_without_policy" {
						name              = "%s"
						enabled           = true
						description       = "initial description"
						expression        = "open.file.name == \"etc/shadow/password\""
						product_tags      = ["compliance_framework:HIPAA"]
					}
					`, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExistsWithoutPolicyID(providers.frameworkProvider, resourceName),
					checkCSMThreatsAgentRuleContent(resourceName, agentRuleName, "initial description", "open.file.name == \"etc/shadow/password\"", "compliance_framework:HIPAA"),
				),
			},
			// update the description
			{
				Config: fmt.Sprintf(`
					resource "datadog_csm_threats_agent_rule" "agent_rule_without_policy" {
						name              = "%s"
						enabled           = true
						description       = "updated description"
						expression        = "open.file.name == \"etc/shadow/password\""
						product_tags      = ["compliance_framework:HIPAA"]
					}
					`, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExistsWithoutPolicyID(providers.frameworkProvider, resourceName),
					checkCSMThreatsAgentRuleContent(resourceName, agentRuleName, "updated description", "open.file.name == \"etc/shadow/password\"", "compliance_framework:HIPAA"),
				),
			},
		},
	})
}

func TestAccCSMThreatsAgentRule_CreateAndUpdateWithActions(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	agentRuleName := uniqueAgentRuleName(ctx)
	resourceName := "datadog_csm_threats_agent_rule.agent_rule_with_actions"

	policyName := uniqueAgentRuleName(ctx)
	policyConfig := fmt.Sprintf(`
	resource "datadog_csm_threats_policy" "policy_for_actions_test" {
		name              = "%s"
		enabled           = true
		description       = "policy for actions test"
		tags              = ["host_name:test_host"]
	}
	`, policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsAgentRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Create a policy to have at least one
				Config: policyConfig,
				Check:  testAccCheckCSMThreatsPolicyExists(providers.frameworkProvider, "datadog_csm_threats_policy.policy_for_actions_test"),
			},
			{
				Config: fmt.Sprintf(`
				%s
				resource "datadog_csm_threats_agent_rule" "agent_rule_with_actions" {
					name              = "%s"
                    policy_id         = datadog_csm_threats_policy.policy_for_actions_test.id
					enabled           = true
					description       = "rule with actions"
					expression 		  = "open.file.name == \"etc/shadow/password\""
					product_tags      = ["compliance_framework:PCI-DSS"]
					actions = [
						{
							set = {
								name   = "my_action_test"
								field  = "test_field"
								value  = ""
								append = false
								scope  = ""
								ttl    = 0
								size   = 0
							}
						}
					]
				}
				`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", agentRuleName),
					resource.TestCheckResourceAttr(resourceName, "description", "rule with actions"),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.name", "my_action_test"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.field", "test_field"),
				),
			},
			// Update description while keeping actions the same
			{
				Config: fmt.Sprintf(`
				%s
				resource "datadog_csm_threats_agent_rule" "agent_rule_with_actions" {
					name              = "%s"
                    policy_id         = datadog_csm_threats_policy.policy_for_actions_test.id
					enabled           = true
					description       = "updated rule with actions"
					expression 		  = "open.file.name == \"etc/shadow/password\""
					product_tags      = ["compliance_framework:ISO-27799"]
					actions = [
						{
							set = {
								name   = "my_action_test"
								field  = "test_field"
								value  = ""
								append = false
								scope  = ""
								ttl    = 0
								size   = 0
							}
						}
					]
				}
				`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", agentRuleName),
					resource.TestCheckResourceAttr(resourceName, "description", "updated rule with actions"),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.name", "my_action_test"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.field", "test_field"),
				),
			},
		},
	})
}

func TestAccCSMThreatsAgentRule_CreateWithActionsWithoutValue(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	agentRuleName := uniqueAgentRuleName(ctx)
	resourceName := "datadog_csm_threats_agent_rule.agent_rule_without_value"

	policyName := uniqueAgentRuleName(ctx)
	policyConfig := fmt.Sprintf(`
	resource "datadog_csm_threats_policy" "policy_for_no_value_test" {
		name              = "%s"
		enabled           = true
		description       = "policy for no value test"
		tags              = ["host_name:test_host"]
	}
	`, policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsAgentRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Create a policy to have at least one
				Config: policyConfig,
				Check:  testAccCheckCSMThreatsPolicyExists(providers.frameworkProvider, "datadog_csm_threats_policy.policy_for_no_value_test"),
			},
			{
				Config: fmt.Sprintf(`
				%s
				resource "datadog_csm_threats_agent_rule" "agent_rule_without_value" {
					name              = "%s"
                    policy_id         = datadog_csm_threats_policy.policy_for_no_value_test.id
					enabled           = true
					description       = "rule with actions without value"
					expression 		  = "open.file.name == \"etc/shadow/password\""
					product_tags      = ["compliance_framework:PCI-DSS"]
					actions = [
						{
							set = {
								name   = "my_action_hhrrr"
								field  = "process.name"
								value  = ""
								append = true
								scope  = "container"
								ttl    = 60
								size   = 10
							}
						}
					]
				}
				`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", agentRuleName),
					resource.TestCheckResourceAttr(resourceName, "description", "rule with actions without value"),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.name", "my_action_hhrrr"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.field", "process.name"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.append", "true"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.scope", "container"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.ttl", "60"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.set.size", "10"),
				),
			},
		},
	})
}

func testAccCheckCSMThreatsAgentRuleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in the state %s", resourceName, s.RootModule().Resources)
		}

		if resource.Type != "datadog_csm_threats_agent_rule" {
			return fmt.Errorf("resource %s is not of type datadog_csm_threats_agent_rule, found %s instead", resourceName, resource.Type)
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

func testAccCheckCSMThreatsAgentRuleExistsWithoutPolicyID(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in the state", resourceName)
		}

		if resource.Type != "datadog_csm_threats_agent_rule" {
			return fmt.Errorf("resource %s is not of type datadog_csm_threats_agent_rule", resourceName)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		_, _, err := apiInstances.GetCSMThreatsApiV2().GetCSMThreatsAgentRule(auth, resource.Primary.ID)
		if err != nil {
			return fmt.Errorf("received an error retrieving agent rule: %s", err)
		}

		return nil
	}
}

func testAccCheckCSMThreatsAgentRuleDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_csm_threats_agent_rule" {
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

func checkCSMThreatsAgentRuleContent(resourceName string, name string, description string, expression string, product_tags string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
		resource.TestCheckResourceAttr(resourceName, "expression", expression),
		resource.TestCheckResourceAttr(resourceName, "product_tags.#", "1"),
		resource.TestCheckTypeSetElemAttr(resourceName, "product_tags.*", product_tags),
	)
}
