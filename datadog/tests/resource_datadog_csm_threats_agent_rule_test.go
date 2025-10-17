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
					actions {
						set {
							name   = "test_action"
							field  = "exec.file.path"
							append = false
							scope  = "process"
						}
						hash {}
					}
				}`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, resourceName),
					checkCSMThreatsAgentRuleContent(
						resourceName,
						agentRuleName,
						"im a rule",
						"open.file.name == \"etc/shadow/password\"",
						"compliance_framework:PCI-DSS",
						"test_action",
						"exec.file.path",
						"process",
					),
				),
			},
			// Update description and actions
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
					actions {
						set {
							name   = "updated_action"
							value  = "new_value"
							append = true
							scope  = "container"
						}
						hash {}
					}
				}`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, resourceName),
					checkCSMThreatsAgentRuleContent(
						resourceName,
						agentRuleName,
						"updated agent rule for terraform provider test",
						"open.file.name == \"etc/shadow/password\"",
						"compliance_framework:ISO-27799",
						"updated_action",
						"new_value",
						"container",
					),
				),
			},
			// Update actions
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
					actions {
						set {
							name   = "updated_action"
							expression  = "\"value_$${builtins.uuid4}\""
							scope  = "container"
							inherited = true
							default_value = "abc"
						}
					}
				}`, policyConfig, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, resourceName),
					checkCSMThreatsAgentRuleContent(
						resourceName,
						agentRuleName,
						"updated agent rule for terraform provider test",
						"open.file.name == \"etc/shadow/password\"",
						"compliance_framework:ISO-27799",
						"updated_action",
						"new_value",
						"container",
					),
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

func checkCSMThreatsAgentRuleContent(resourceName string, name string, description string, expression string, product_tags string, action_name string, action_value_source string, action_scope string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
		resource.TestCheckResourceAttr(resourceName, "expression", expression),
		resource.TestCheckResourceAttr(resourceName, "product_tags.#", "1"),
		resource.TestCheckTypeSetElemAttr(resourceName, "product_tags.*", product_tags),
		resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "actions.0.set.name", action_name),
		resource.TestCheckResourceAttr(resourceName, "actions.0.set.scope", action_scope),
		resource.TestCheckResourceAttr(resourceName, "actions.0.hash.%", "0"),
		func(s *terraform.State) error {
			r := s.RootModule().Resources[resourceName]
			if r == nil {
				return fmt.Errorf("resource not found")
			}

			// Check either value or field is set (but not both)
			value := r.Primary.Attributes["actions.0.set.value"]
			field := r.Primary.Attributes["actions.0.set.field"]
			expression := r.Primary.Attributes["actions.0.set.expression"]

			if value == action_value_source {
				if field != "" {
					return fmt.Errorf("both value and field are set")
				}
			} else if field == action_value_source {
				if value != "" {
					return fmt.Errorf("both value and field are set")
				}
			} else if expression == action_value_source {
				if value != "" {
					return fmt.Errorf("both value and expression are set")
				}
			} else {
				return fmt.Errorf("neither value nor field matches expected value")
			}

			return nil
		},
	)
}
