package test

import (
	"context"
	"encoding/json"
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
		ProtoV6ProviderFactories: accProviders,
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
	rule := map[string]interface{}{
		"name":      name,
		"isEnabled": false,
		"type":      "log_detection",
		"message":   "Test rule triggered",
		"tags":      []string{"test:tag"},
		"cases": []map[string]interface{}{
			{
				"status":        "info",
				"notifications": []string{"@slack-test"},
				"condition":     "a > 0",
			},
		},
		"options": map[string]interface{}{
			"evaluationWindow":  60,
			"keepAlive":         60,
			"maxSignalDuration": 60,
			"detectionMethod":   "threshold",
		},
		"queries": []map[string]interface{}{
			{
				"query":          "source:test",
				"aggregation":    "count",
				"groupByFields":  []string{"host"},
				"distinctFields": []string{},
				"name":           "a",
				"dataSource":     "logs",
			},
		},
	}
	b, _ := json.Marshal(rule)
	return string(b)
}

func testAccSecurityMonitoringRuleJSONUpdated(name string) string {
	rule := map[string]interface{}{
		"name":      fmt.Sprintf("%s - updated", name),
		"isEnabled": true,
		"type":      "log_detection",
		"message":   "Test rule triggered (updated)",
		"tags":      []string{"env:test", "test:tag"},
		"cases": []map[string]interface{}{
			{
				"status":        "high",
				"notifications": []string{"@slack-test"},
				"condition":     "a > 10",
			},
		},
		"options": map[string]interface{}{
			"evaluationWindow":  60,
			"keepAlive":         60,
			"maxSignalDuration": 60,
			"detectionMethod":   "threshold",
		},
		"queries": []map[string]interface{}{
			{
				"query":          "source:test-updated",
				"aggregation":    "count",
				"groupByFields":  []string{"host", "service"},
				"distinctFields": []string{},
				"name":           "a",
				"dataSource":     "logs",
			},
		},
	}
	b, _ := json.Marshal(rule)
	return string(b)
}
