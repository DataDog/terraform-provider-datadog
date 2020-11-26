package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

const tfSecurityRuleName = "datadog_security_monitoring_rule.acceptance_test"

func TestAccDatadogSecurityMonitoringRule_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	ruleName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName),
				Check:  testAccCheckDatadogSecurityMonitorCreatedCheck(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringUpdatedConfig(ruleName),
				Check:  testAccCheckDatadogSecurityMonitoringUpdateCheck(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringEnabledDefaultConfig(ruleName),
				Check:  testAccCheckDatadogSecurityMonitoringEnabledDefaultCheck(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringRule_OnlyRequiredFields(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	ruleName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringCreatedRequiredConfig(ruleName),
				Check:  testAccCheckDatadogSecurityMonitorCreatedRequiredCheck(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringUpdatedConfig(ruleName),
				Check:  testAccCheckDatadogSecurityMonitoringUpdateCheck(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringRule_Import(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	ruleName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringCreatedRequiredConfig(ruleName),
			},
			{
				ResourceName:            tfSecurityRuleName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
				Check:                   testAccCheckDatadogSecurityMonitorCreatedRequiredCheck(accProvider, ruleName),
			},
		},
	})
}

func testAccCheckDatadogSecurityMonitoringCreatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
    name = "%s"
    message = "acceptance rule triggered"
    enabled = false

    query { 
        name = "first"
        query = "does not really match much"
        aggregation = "count"
        group_by_fields = ["host"]
    }

    query { 
        name = "second"
        query = "does not really match much either"
        aggregation = "cardinality"
        distinct_fields = ["@orgId"]
        group_by_fields = ["host"]
        metric = "@network.bytes_read"
    }

	case {
        name = "high case"
        status = "high"
        condition = "first > 3 || second > 10"
        notifications = ["@user"]
    }

    case {
		name = "warning case"
        status = "medium"
        condition = "first > 0 || second > 0"
    }

    options {
        evaluation_window = 300
        keep_alive = 600
        max_signal_duration = 900
    }

    tags = ["i:tomato", "u:tomato"]
}
`, name)
}

func testAccCheckDatadogSecurityMonitorCreatedCheck(accProvider *schema.Provider, ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.name", "first"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.query", "does not really match much"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.aggregation", "count"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "host"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.1.name", "second"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.1.query", "does not really match much either"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.1.aggregation", "cardinality"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.1.distinct_fields.0", "@orgId"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.1.group_by_fields.0", "host"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.1.metric", "@network.bytes_read"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", "high case"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "high"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.condition", "first > 3 || second > 10"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.1.name", "warning case"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.1.status", "medium"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.1.condition", "first > 0 || second > 0"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "u:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
    name = "%s - updated"
    message = "acceptance rule triggered (updated)"
    enabled = true

    query { 
        name = "first_updated"
        query = "does not really match much (updated)"
        aggregation = "cardinality"
        distinct_fields = ["@orgId"]
        group_by_fields = ["service"]
        metric = "@network.bytes_read"
    }

	case {
        name = "high case (updated)"
        status = "medium"
        condition = "first_updated > 3"
        notifications = ["@user"]
    }

    case {
		name = "warning case (updated)"
        status = "high"
        condition = "first_updated > 0"
    }

    options {
        evaluation_window = 60
        keep_alive = 300
        max_signal_duration = 600
    }

    tags = ["u:tomato", "i:tomato"]
}
`, name)
}

func testAccCheckDatadogSecurityMonitoringUpdateCheck(accProvider *schema.Provider, ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName+" - updated"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "true"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.name", "first_updated"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.query", "does not really match much (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.aggregation", "cardinality"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.distinct_fields.0", "@orgId"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "service"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.metric", "@network.bytes_read"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "medium"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.condition", "first_updated > 3"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.1.name", "warning case (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.1.status", "high"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.1.condition", "first_updated > 0"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.evaluation_window", "60"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "300"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "u:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "i:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringEnabledDefaultConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
    name = "%s - updated"
    message = "acceptance rule triggered (updated)"

    query { 
        name = "first_updated"
        query = "does not really match much (updated)"
        aggregation = "cardinality"
        distinct_fields = ["@orgId"]
        group_by_fields = ["service"]
        metric = "@network.bytes_read"
    }

	case {
        name = "high case (updated)"
        status = "medium"
        condition = "first_updated > 3"
        notifications = ["@user"]
    }

    case {
		name = "warning case (updated)"
        status = "high"
        condition = "first_updated > 0"
    }

    options {
        evaluation_window = 60
        keep_alive = 300
        max_signal_duration = 600
    }

    tags = ["u:tomato", "i:tomato"]
}
`, name)
}

func testAccCheckDatadogSecurityMonitoringEnabledDefaultCheck(accProvider *schema.Provider, ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName+" - updated"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "true"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.name", "first_updated"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.query", "does not really match much (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.aggregation", "cardinality"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.distinct_fields.0", "@orgId"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "service"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.metric", "@network.bytes_read"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "medium"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.condition", "first_updated > 3"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.1.name", "warning case (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.1.status", "high"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.1.condition", "first_updated > 0"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.evaluation_window", "60"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "300"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "u:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "i:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringCreatedRequiredConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
    name = "%s"
    message = "acceptance rule triggered"

    query { 
        query = "does not really match much"
        aggregation = "count"
        group_by_fields = ["host"]
    }

	case {
        status = "high"
		condition = "a > 0"
    }

    options {
        evaluation_window = 300
        keep_alive = 600
        max_signal_duration = 900
    }
}
`, name)
}

func testAccCheckDatadogSecurityMonitorCreatedRequiredCheck(accProvider *schema.Provider, ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.query", "does not really match much"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.aggregation", "count"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "host"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "high"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.condition", "a > 0"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
	)
}

func testAccCheckDatadogSecurityMonitoringRuleExists(provider *schema.Provider, rule string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := provider.Meta().(*ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2

		for _, rule := range s.RootModule().Resources {
			_, _, err := client.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, rule.Primary.ID).Execute()
			if err != nil {
				return fmt.Errorf("received an error retrieving security monitoring rule: %s", err)
			}
		}
		return nil
	}
}

func testAccCheckDatadogSecurityMonitoringRuleDestroy(provider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := provider.Meta().(*ProviderConfiguration)
		authV2 := providerConf.AuthV2
		client := providerConf.DatadogClientV2

		for _, rule := range s.RootModule().Resources {
			_, httpResponse, err := client.SecurityMonitoringApi.GetSecurityMonitoringRule(authV2, rule.Primary.ID).Execute()
			if err != nil {
				if httpResponse != nil && httpResponse.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error deleting security monitoring rule: %s", err)
			}
			return fmt.Errorf("monitor still exists")
		}
		return nil
	}

}
