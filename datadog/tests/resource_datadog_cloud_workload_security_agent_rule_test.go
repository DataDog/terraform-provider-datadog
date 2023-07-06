package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

const tfAgentRuleName = "datadog_cloud_workload_security_agent_rule.acceptance_test"

func TestAccDatadogCloudWorkloadSecurityAgentRule(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	agentRuleName := strings.Replace(uniqueEntityName(ctx, t), "-", "_", -1)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCloudWorkloadSecurityAgentRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudWorkloadSecurityAgentRuleCreated(agentRuleName),
				Check:  testAccCheckDatadogCloudWorkloadSecurityAgentRuleCreatedCheck(accProvider, agentRuleName),
			},
			{
				Config: testAccCheckDatadogCloudWorkloadSecurityAgentRuleUpdated(agentRuleName),
				Check:  testAccCheckDatadogCloudWorkloadSecurityAgentRuleUpdatedCheck(accProvider, agentRuleName),
			},
		},
	})
}

func testAccCheckDatadogCloudWorkloadSecurityAgentRuleCreated(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_workload_security_agent_rule" "acceptance_test" {
    name = "%s"
    description = "an agent rule"
    enabled = "true"
	expression = "exec.file.name == \"java\""
}
`, name)
}

func testAccCheckDatadogCloudWorkloadSecurityAgentRuleCreatedCheck(accProvider func() (*schema.Provider, error), agentRuleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudWorkloadSecurityAgentRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfAgentRuleName, "name", agentRuleName),
		resource.TestCheckResourceAttr(
			tfAgentRuleName, "description", "an agent rule"),
		resource.TestCheckResourceAttr(
			tfAgentRuleName, "enabled", "true"),
		resource.TestCheckResourceAttr(
			tfAgentRuleName, "expression", "exec.file.name == \"java\""),
	)
}

func testAccCheckDatadogCloudWorkloadSecurityAgentRuleUpdated(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_workload_security_agent_rule" "acceptance_test" {
    name = "%s"
    description = "a new agent rule"
    enabled = "false"
	expression = "exec.file.name == \"go\""
}
`, name)
}

func testAccCheckDatadogCloudWorkloadSecurityAgentRuleUpdatedCheck(accProvider func() (*schema.Provider, error), agentRuleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudWorkloadSecurityAgentRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfAgentRuleName, "name", agentRuleName),
		resource.TestCheckResourceAttr(
			tfAgentRuleName, "description", "a new agent rule"),
		resource.TestCheckResourceAttr(
			tfAgentRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfAgentRuleName, "expression", "exec.file.name == \"go\""),
	)
}

func testAccCheckDatadogCloudWorkloadSecurityAgentRuleExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, agentRule := range s.RootModule().Resources {
			_, _, err := apiInstances.GetCloudWorkloadSecurityApiV2().GetCloudWorkloadSecurityAgentRule(auth, agentRule.Primary.ID)
			if err != nil {
				return fmt.Errorf("received an error retrieving cloud workload security agent rule: %s", err)
			}
		}
		return nil
	}
}

func testAccCheckDatadogCloudWorkloadSecurityAgentRuleDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_cloud_workload_security_agent_rule" {
				_, httpResponse, err := apiInstances.GetCloudWorkloadSecurityApiV2().GetCloudWorkloadSecurityAgentRule(auth, resource.Primary.ID)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == 404 {
						continue
					}
					return fmt.Errorf("received an error deleting cloud workload security agent rule: %s", err)
				}
				return fmt.Errorf("cloud workload security agent rule still exists")
			}
		}
		return nil
	}

}
