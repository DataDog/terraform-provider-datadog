package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

const tfSecurityRuleJSONName = "datadog_security_monitoring_rule_json.acceptance_test"

func TestAccDatadogSecurityMonitoringRuleJSON_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityMonitoringRuleJSONDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringRuleJSONConfig(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityMonitoringRuleJSONExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						tfSecurityRuleJSONName, "json", testAccSecurityMonitoringRuleJSON(ruleName),
					),
				),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringRuleJSONUpdatedConfig(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityMonitoringRuleJSONExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						tfSecurityRuleJSONName, "json", testAccSecurityMonitoringRuleJSONUpdated(ruleName),
					),
				),
			},
		},
	})
}

func testAccCheckDatadogSecurityMonitoringRuleJSONExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceRule, ok := s.RootModule().Resources[tfSecurityRuleJSONName]
		if !ok {
			return fmt.Errorf("security monitoring rule json not found in state: %s", tfSecurityRuleJSONName)
		}

		if resourceRule.Primary == nil {
			return fmt.Errorf("security monitoring rule json resource has no primary instance")
		}

		_, httpResp, err := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(
			accProvider.Auth,
			resourceRule.Primary.ID,
		)

		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return fmt.Errorf("security monitoring rule json not found")
			}
			return fmt.Errorf("received an error retrieving security monitoring rule json %s", err)
		}

		return nil
	}
}

func testAccCheckDatadogSecurityMonitoringRuleJSONDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, resource := range s.RootModule().Resources {
			if resource.Type != "datadog_security_monitoring_rule_json" {
				continue
			}

			_, httpResp, err := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(
				accProvider.Auth,
				resource.Primary.ID,
			)

			if err == nil {
				return fmt.Errorf("security monitoring rule json still exists")
			}

			if httpResp == nil || httpResp.StatusCode != 404 {
				return fmt.Errorf("received an error when expecting a 404: %s", err)
			}
		}

		return nil
	}
}

func testAccCheckDatadogSecurityMonitoringRuleJSONConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule_json" "acceptance_test" {
	json = jsonencode(%s)
}`, testAccSecurityMonitoringRuleJSON(name))
}

func testAccCheckDatadogSecurityMonitoringRuleJSONUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule_json" "acceptance_test" {
	json = jsonencode(%s)
}`, testAccSecurityMonitoringRuleJSONUpdated(name))
}

func testAccSecurityMonitoringRuleJSON(name string) string {
	return fmt.Sprintf(`{
		"name": "%s",
		"isEnabled": false,
		"type": "log_detection",
		"message": "Test rule triggered",
		"tags": ["test:tag"],
		"cases": [{
			"status": "info",
			"notifications": ["@slack-test"],
			"condition": "a > 0"
		}],
		"options": {
			"evaluationWindow": 300,
			"keepAlive": 600,
			"maxSignalDuration": 900,
			"detectionMethod": "threshold"
		},
		"queries": [{
			"query": "source:test",
			"aggregation": "count",
			"groupByFields": ["host"],
			"distinctFields": [],
			"name": "a",
			"dataSource": "logs"
		}]
	}`, name)
}

func testAccSecurityMonitoringRuleJSONUpdated(name string) string {
	return fmt.Sprintf(`{
		"name": "%s - updated",
		"isEnabled": true,
		"type": "log_detection",
		"message": "Test rule triggered (updated)",
		"tags": ["test:tag", "env:test"],
		"cases": [{
			"status": "high",
			"notifications": ["@slack-test"],
			"condition": "a > 10"
		}],
		"options": {
			"evaluationWindow": 600,
			"keepAlive": 900,
			"maxSignalDuration": 1200,
			"detectionMethod": "threshold"
		},
		"queries": [{
			"query": "source:test-updated",
			"aggregation": "count",
			"groupByFields": ["host", "service"],
			"distinctFields": [],
			"name": "a",
			"dataSource": "logs"
		}]
	}`, name)
}
