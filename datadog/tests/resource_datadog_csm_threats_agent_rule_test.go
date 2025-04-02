package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// Create an agent rule and update its description
func TestAccCSMThreatsAgentRule_CreateAndUpdate(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	agentRuleName := uniqueAgentRuleName(ctx)
	resourceName := "datadog_csm_threats_agent_rule.agent_rule_test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsAgentRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "datadog_csm_threats_agent_rule" "agent_rule_test" {
					name              = "%s"
					enabled           = true
					description       = "im a rule"
					expression 		  = "open.file.name == \"etc/shadow/password\""
				}
				`, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, "datadog_csm_threats_agent_rule.agent_rule_test"),
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
				resource "datadog_csm_threats_agent_rule" "agent_rule_test" {
					name              = "%s"
					enabled           = true
					description       = "updated agent rule for terraform provider test"
					expression 		  = "open.file.name == \"etc/shadow/password\""
				}
				`, agentRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsAgentRuleExists(providers.frameworkProvider, resourceName),
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

func checkCSMThreatsAgentRuleContent(resourceName string, name string, description string, expression string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
		resource.TestCheckResourceAttr(resourceName, "expression", expression),
	)
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
				_, httpResponse, err := apiInstances.GetCSMThreatsApiV2().GetCSMThreatsAgentRule(auth, resource.Primary.ID)
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
