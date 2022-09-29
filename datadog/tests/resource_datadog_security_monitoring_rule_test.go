package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

const tfSecurityRuleName = "datadog_security_monitoring_rule.acceptance_test"

func TestAccDatadogSecurityMonitoringRule_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
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

func TestAccDatadogSecurityMonitoringRule_NewValueRule(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringCreatedConfigNewValueRule(ruleName),
				Check:  testAccCheckDatadogSecurityMonitorCreatedCheckNewValueRule(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringUpdatedConfigNewValueRule(ruleName),
				Check:  testAccCheckDatadogSecurityMonitoringUpdateCheckNewValueRule(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringRule_ImpossibleTravelRule(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringCreatedConfigImpossibleTravelRule(ruleName),
				Check:  testAccCheckDatadogSecurityMonitorCreatedCheckImpossibleTravelRule(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringUpdatedConfigImpossibleTravelRule(ruleName),
				Check:  testAccCheckDatadogSecurityMonitorUpdatedCheckImpossibleTravelRule(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringRule_CwsRule(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringCreatedConfigCwsRule(ruleName),
				Check:  testAccCheckDatadogSecurityMonitoringCreatedCheckCwsRule(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringUpdatedConfigCwsRule(ruleName),
				Check:  testAccCheckDatadogSecurityMonitoringUpdateCheckCwsRule(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringRule_OnlyRequiredFields(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringCreatedRequiredConfig(ruleName),
			},
			{
				ResourceName:      tfSecurityRuleName,
				ImportState:       true,
				ImportStateVerify: true,
				Check:             testAccCheckDatadogSecurityMonitorCreatedRequiredCheck(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringRule_SignalCorrelation(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityMonitoringCreatedSignalCorrelationConfig(ruleName),
				Check:  testAccCheckDatadogSecurityMonitorCreatedSignalCorrelationCheck(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogSecurityMonitoringUpdatedSignalCorrelationConfig(ruleName),
				Check:  testAccCheckDatadogSecurityMonitoringUpdateSignalCorrelationCheck(accProvider, ruleName),
			},
		},
	})
}

func testAccCheckDatadogSecurityMonitoringCreatedConfig(name string) string {
	return testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(name, "")
}

func testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(name string, id string) string {
	suffix := id
	if suffix != "" {
		suffix = fmt.Sprintf("_%s", suffix)
	}
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test%s" {
	standard_rule {
		name = "%s"
		message = "acceptance rule triggered"
		enabled = false
		has_extended_title = true
	
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
		}
	
		query {
			name = "third"
			query = "does not really match much either"
			aggregation = "sum"
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
	
		case {
			name = "low case"
			status = "low"
			condition = "third > 9000"
		}
	
		options {
			evaluation_window = 300
			keep_alive = 600
			max_signal_duration = 900
			decrease_criticality_based_on_env = true
		}
	
		filter {
			query = "does not really suppress"
			action = "suppress"
		}
	
		filter {
			query = "does not really require neither"
			action = "require"
		}
	
		tags = ["i:tomato", "u:tomato"]
	}
}
`, suffix, name)
}

func testAccCheckDatadogSecurityMonitorCreatedCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return testAccCheckDatadogSecurityMonitorCreatedCheckWithId(accProvider, ruleName, "")
}

func testCheckResourceAttrStandardRule(name string, key string, value string) resource.TestCheckFunc {
	fullKey := fmt.Sprintf("standard_rule.0.%s", key)
	return resource.TestCheckResourceAttr(name, fullKey, value)
}

func testCheckResourceAttrSignalRule(name string, key string, value string) resource.TestCheckFunc {
	fullKey := fmt.Sprintf("signal_rule.0.%s", key)
	return resource.TestCheckResourceAttr(name, fullKey, value)
}

func testAccCheckDatadogSecurityMonitorCreatedCheckWithId(accProvider func() (*schema.Provider, error), ruleName string, id string) resource.TestCheckFunc {
	tfSecurityRuleNameWithId := tfSecurityRuleName
	if id != "" {
		tfSecurityRuleNameWithId = fmt.Sprintf("%s_%s", tfSecurityRuleNameWithId, id)
	}
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleNameWithId),
		testCheckResourceAttrStandardRule(tfSecurityRuleNameWithId, "name", ruleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "message", "acceptance rule triggered"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "enabled", "false"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "has_extended_title", "true"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.0.name", "first"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.0.query", "does not really match much"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.0.aggregation", "count"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.0.group_by_fields.0", "host"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.1.name", "second"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.1.query", "does not really match much either"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.1.aggregation", "cardinality"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.1.distinct_fields.0", "@orgId"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.1.group_by_fields.0", "host"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.2.name", "third"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.2.query", "does not really match much either"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.2.aggregation", "sum"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.2.group_by_fields.0", "host"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "query.2.metric", "@network.bytes_read"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.0.name", "high case"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.0.status", "high"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.0.condition", "first > 3 || second > 10"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.0.notifications.0", "@user"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.1.name", "warning case"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.1.status", "medium"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.1.condition", "first > 0 || second > 0"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.2.name", "low case"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.2.status", "low"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "case.2.condition", "third > 9000"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "options.0.evaluation_window", "300"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "options.0.keep_alive", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "options.0.max_signal_duration", "900"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "options.0.decrease_criticality_based_on_env", "true"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "filter.0.action", "suppress"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "filter.0.query", "does not really suppress"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "filter.1.action", "require"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "filter.1.query", "does not really require neither"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "tags.0", "i:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleNameWithId, "tags.1", "u:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringCreatedConfigNewValueRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
	standard_rule {
		name = "%s"
		message = "acceptance rule triggered"
		enabled = false
	
		query {
			name = "first"
			query = "does not really match much"
			aggregation = "new_value"
			metric = "@value"
			group_by_fields = ["host"]
		}
	
		case {
			name = ""
			status = "high"
			notifications = ["@user"]
		}
	
		options {
			detection_method = "new_value"
			keep_alive = 600
			max_signal_duration = 900
			new_value_options {
				forget_after = 7
				learning_duration = 1
			}
		}
	
		tags = ["i:tomato", "u:tomato"]
	}
}
`, name)
}

func testAccCheckDatadogSecurityMonitorCreatedCheckNewValueRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "name", ruleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "message", "acceptance rule triggered"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "enabled", "false"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.name", "first"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.query", "does not really match much"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.aggregation", "new_value"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.group_by_fields.0", "host"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.name", ""),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.status", "high"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.detection_method", "new_value"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.new_value_options.0.forget_after", "7"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_method", "duration"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_duration", "1"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_threshold", "0"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.1", "u:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringCreatedConfigImpossibleTravelRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
    standard_rule {
		name = "%s"
		message = "impossible travel rule triggered"
		enabled = false
	
		query {
			name = "my_query"
			query = "*"
			aggregation = "geo_data"
			metric = "@usr.handle"
			group_by_fields = ["@usr.handle"]
		}
	
		case {
			name = ""
			status = "high"
			notifications = ["@user"]
		}
	
		options {
			detection_method = "impossible_travel"
			keep_alive = 600
			max_signal_duration = 900
			impossible_travel_options {
				baseline_user_locations = true
			}
		}
	
		tags = ["i:tomato", "u:tomato"]
	}
}
`, name)
}

func testAccCheckDatadogSecurityMonitorCreatedCheckImpossibleTravelRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "name", ruleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "message", "impossible travel rule triggered"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "enabled", "false"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.name", "my_query"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.query", "*"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.aggregation", "geo_data"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.group_by_fields.0", "@usr.handle"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.name", ""),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.status", "high"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.detection_method", "impossible_travel"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.impossible_travel_options.0.baseline_user_locations", "true"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.1", "u:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedConfigImpossibleTravelRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
	standard_rule {
		name = "%s"
		message = "impossible travel rule triggered (updated)"
		enabled = false
	
		query {
			name = "my_updated_query"
			query = "*"
			aggregation = "geo_data"
			metric = "@usr.handle"
			group_by_fields = ["@usr.handle"]
		}
	
		case {
			name = "new case name (updated)"
			status = "high"
			notifications = ["@user"]
		}
	
		options {
			detection_method = "impossible_travel"
			keep_alive = 600
			max_signal_duration = 900
			impossible_travel_options {
				baseline_user_locations = true
			}
		}
	
		tags = ["i:tomato", "u:tomato"]
	}
}
`, name)
}

func testAccCheckDatadogSecurityMonitorUpdatedCheckImpossibleTravelRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "name", ruleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "message", "impossible travel rule triggered (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "enabled", "false"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.name", "my_updated_query"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.query", "*"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.aggregation", "geo_data"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.group_by_fields.0", "@usr.handle"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.name", "new case name (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.status", "high"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.detection_method", "impossible_travel"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.impossible_travel_options.0.baseline_user_locations", "true"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.1", "u:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringCreatedConfigCwsRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
	standard_rule {
		name = "%s"
		message = "acceptance rule triggered"
		enabled = false
	
		query {
			name = "first"
			query = "@agent.rule_id:(%s_random_id OR random_id)"
			aggregation = "count"
			group_by_fields = ["host"]
		}
	
		case {
			name = "high case"
			status = "high"
			condition = "first > 3"
		}
	
		options {
			detection_method = "threshold"
			evaluation_window = 300
			keep_alive = 600
			max_signal_duration = 900
		}
	
		tags = ["i:tomato", "u:tomato"]
	
		type = "workload_security"
	}
}
`, name, strings.Replace(name, "-", "_", -1))
}

func testAccCheckDatadogSecurityMonitoringCreatedCheckCwsRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "name", ruleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "message", "acceptance rule triggered"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "enabled", "false"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.name", "first"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.query", fmt.Sprintf("@agent.rule_id:(%s_random_id OR random_id)", strings.Replace(ruleName, "-", "_", -1))),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.aggregation", "count"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.group_by_fields.0", "host"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.name", "high case"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.status", "high"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.condition", "first > 3"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.detection_method", "threshold"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.1", "u:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "type", "workload_security"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
	standard_rule {
		name = "%s - updated"
		message = "acceptance rule triggered (updated)"
		enabled = true
		has_extended_title = false
	
		query {
			name = "first_updated"
			query = "does not really match much (updated)"
			aggregation = "cardinality"
			distinct_fields = ["@orgId"]
			group_by_fields = ["service"]
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
	
		filter {
			query = "does not really suppress (updated)"
			action = "suppress"
		}
	
		tags = ["u:tomato", "i:tomato"]
	}
}
`, name)
}

func testAccCheckDatadogSecurityMonitoringUpdateCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "name", ruleName+" - updated"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "enabled", "true"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "has_extended_title", "false"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.name", "first_updated"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.query", "does not really match much (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.aggregation", "cardinality"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.distinct_fields.0", "@orgId"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.group_by_fields.0", "service"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.status", "medium"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.condition", "first_updated > 3"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.1.name", "warning case (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.1.status", "high"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.1.condition", "first_updated > 0"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.evaluation_window", "60"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.keep_alive", "300"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.decrease_criticality_based_on_env", "false"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "filter.0.action", "suppress"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "filter.0.query", "does not really suppress (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.0", "u:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.1", "i:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedConfigNewValueRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
	standard_rule {
		name = "%s - updated"
		message = "acceptance rule triggered (updated)"
		enabled = true
	
		query {
			name = "first"
			query = "does not really match much (updated)"
			aggregation = "new_value"
			group_by_fields = ["service"]
			metric = "@network.bytes_read"
		}
	
		case {
			name = "high case (updated)"
			status = "medium"
			condition = ""
			notifications = ["@user"]
		}
	
		 options {
			detection_method = "new_value"
			keep_alive = 600
			max_signal_duration = 900
			new_value_options {
				forget_after = 1
				learning_duration = 0
			}
		}
	
		tags = ["u:tomato", "i:tomato"]
	}
}
`, name)
}

func testAccCheckDatadogSecurityMonitoringUpdateCheckNewValueRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "name", ruleName+" - updated"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "enabled", "true"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.name", "first"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.query", "does not really match much (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.aggregation", "new_value"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.group_by_fields.0", "service"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.metric", "@network.bytes_read"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.status", "medium"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.detection_method", "new_value"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.new_value_options.0.forget_after", "1"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_method", "duration"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_duration", "0"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_threshold", "0"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.0", "u:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.1", "i:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedConfigCwsRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
	standard_rule {
		name = "%s"
		message = "acceptance rule triggered (updated)"
		enabled = true
	
		query {
			name = "first"
			query = "@agent.rule_id:(%s_random_id OR random_id)"
			aggregation = "count"
			group_by_fields = ["service"]
		}
	
		case {
			name = "high case (updated)"
			status = "medium"
			condition = "first > 10"
			notifications = ["@user"]
		}
	
		 options {
			detection_method = "threshold"
			evaluation_window = 300
			keep_alive = 600
			max_signal_duration = 900
		}
	
		tags = ["u:tomato", "i:tomato"]
	
		type = "workload_security"
	}
}
`, name, strings.Replace(name, "-", "_", -1))
}

func testAccCheckDatadogSecurityMonitoringUpdateCheckCwsRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "name", ruleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "enabled", "true"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.name", "first"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.query", fmt.Sprintf("@agent.rule_id:(%s_random_id OR random_id)", strings.Replace(ruleName, "-", "_", -1))),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.aggregation", "count"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.group_by_fields.0", "service"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.status", "medium"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.condition", "first > 10"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.detection_method", "threshold"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.0", "u:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.1", "i:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "type", "workload_security"),
	)
}

func testAccCheckDatadogSecurityMonitoringEnabledDefaultConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
	standard_rule {
		name = "%s - updated"
		message = "acceptance rule triggered (updated)"
	
		query {
			name = "first_updated"
			query = "does not really match much (updated)"
			aggregation = "cardinality"
			distinct_fields = ["@orgId"]
			group_by_fields = ["service"]
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
	
		filter {
			query = "does not really suppress (updated)"
			action = "suppress"
		}
	
		tags = ["u:tomato", "i:tomato"]
	}
}
`, name)
}

func testAccCheckDatadogSecurityMonitoringCreatedSignalCorrelationConfig(name string) string {
	logDetectionRule0 := fmt.Sprintf("%s_rule_0", name)
	logDetectionRule1 := fmt.Sprintf("%s_rule_1", name)
	return fmt.Sprintf(`
%s
%s
resource "datadog_security_monitoring_rule" "acceptance_test" {
	signal_rule {
		name = "%s"
		message = "acceptance rule triggered"
		enabled = false
		has_extended_title = true
	
		query {
			name = "first"
			rule_id = "${datadog_security_monitoring_rule.acceptance_test_0.id}"
			aggregation = "event_count"
			correlated_by_fields = ["host"]
		}
	
		query {
			name = "second"
			rule_id = "${datadog_security_monitoring_rule.acceptance_test_1.id}"
			aggregation = "event_count"
			correlated_by_fields = ["host"]
			correlated_query_index = "1"
		}
	
		case {
			name = "high case"
			status = "high"
			condition = "first > 0 && second > 0"
			notifications = ["@user"]
		}
	
		options {
			evaluation_window = 300
			keep_alive = 600
			max_signal_duration = 900
		}
	
		filter {
			query = "does not really suppress"
			action = "suppress"
		}
	
		filter {
			query = "does not really require neither"
			action = "require"
		}
	
		type = "signal_correlation"
	
		tags = ["alert:red", "attack:advanced"]
	}
}
`, testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(logDetectionRule0, "0"),
		testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(logDetectionRule1, "1"),
		name)
}

func testCheckResourceAttrPairSignalCorrelation(name string, queryId int, logDetectionId int) resource.TestCheckFunc {
	logDetectionName := fmt.Sprintf("%s_%d", name, logDetectionId)
	signalQuery := fmt.Sprintf("signal_rule.0.query.%d.rule_id", queryId)
	return resource.TestCheckResourceAttrPair(name, signalQuery, logDetectionName, "id")
}

func testAccCheckDatadogSecurityMonitorCreatedSignalCorrelationCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "name", ruleName),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "message", "acceptance rule triggered"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "enabled", "false"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "has_extended_title", "true"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.0.name", "first"),
		testCheckResourceAttrPairSignalCorrelation(tfSecurityRuleName, 0, 0),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.0.aggregation", "event_count"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.0.correlated_by_fields.0", "host"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.0.correlated_query_index", ""),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.1.name", "second"),
		testCheckResourceAttrPairSignalCorrelation(tfSecurityRuleName, 1, 1),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.1.aggregation", "event_count"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.1.correlated_by_fields.0", "host"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.1.correlated_query_index", "1"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "case.0.name", "high case"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "case.0.status", "high"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "case.0.condition", "first > 0 && second > 0"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "filter.0.action", "suppress"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "filter.0.query", "does not really suppress"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "filter.1.action", "require"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "filter.1.query", "does not really require neither"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "tags.0", "alert:red"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "tags.1", "attack:advanced"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedSignalCorrelationConfig(name string) string {
	logDetectionRule0 := fmt.Sprintf("%s_rule_0", name)
	logDetectionRule1 := fmt.Sprintf("%s_rule_1", name)
	return fmt.Sprintf(`
%s
%s
resource "datadog_security_monitoring_rule" "acceptance_test" {
	signal_rule {
		name = "%s - updated"
		message = "acceptance rule triggered (updated)"
		enabled = true
		has_extended_title = false
	
		query {
			name = "first_updated"
			rule_id = "${datadog_security_monitoring_rule.acceptance_test_0.id}"
			correlated_by_fields = ["service"]
		}
	
		query {
			name = "second_updated"
			rule_id = "${datadog_security_monitoring_rule.acceptance_test_1.id}"
			correlated_by_fields = ["service"]
			correlated_query_index = "0"
		}
	
		case {
			name = "high case (updated)"
			status = "medium"
			condition = "first_updated > 0 && second_updated > 0"
			notifications = ["@user"]
		}
	
		options {
			evaluation_window = 60
			keep_alive = 300
			max_signal_duration = 600
		}
	
		filter {
			query = "does not really suppress (updated)"
			action = "suppress"
		}
	
		tags = ["alert:red", "attack:advanced"]
	}
}
`, testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(logDetectionRule0, "0"),
		testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(logDetectionRule1, "1"),
		name)
}

func testAccCheckDatadogSecurityMonitoringUpdateSignalCorrelationCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "name", ruleName+" - updated"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "enabled", "true"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "has_extended_title", "false"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.0.name", "first_updated"),
		testCheckResourceAttrPairSignalCorrelation(tfSecurityRuleName, 0, 0),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.0.aggregation", "event_count"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.0.correlated_by_fields.0", "service"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.0.correlated_query_index", ""),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.1.name", "second_updated"),
		testCheckResourceAttrPairSignalCorrelation(tfSecurityRuleName, 1, 1),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.1.aggregation", "event_count"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.1.correlated_by_fields.0", "service"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "query.1.correlated_query_index", "0"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "case.0.status", "medium"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "case.0.condition", "first_updated > 0 && second_updated > 0"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "options.0.evaluation_window", "60"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "options.0.keep_alive", "300"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "600"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "filter.0.action", "suppress"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "filter.0.query", "does not really suppress (updated)"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "tags.0", "alert:red"),
		testCheckResourceAttrSignalRule(
			tfSecurityRuleName, "tags.1", "attack:advanced"),
	)
}

func testAccCheckDatadogSecurityMonitoringEnabledDefaultCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "name", ruleName+" - updated"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "enabled", "true"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.name", "first_updated"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.query", "does not really match much (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.aggregation", "cardinality"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.distinct_fields.0", "@orgId"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.group_by_fields.0", "service"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.status", "medium"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.condition", "first_updated > 3"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.1.name", "warning case (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.1.status", "high"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.1.condition", "first_updated > 0"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.evaluation_window", "60"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.keep_alive", "300"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "filter.0.action", "suppress"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "filter.0.query", "does not really suppress (updated)"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.0", "u:tomato"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "tags.1", "i:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringCreatedRequiredConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
	standard_rule {
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
}
`, name)
}

func testAccCheckDatadogSecurityMonitorCreatedRequiredCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "name", ruleName),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "message", "acceptance rule triggered"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.query", "does not really match much"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.aggregation", "count"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "query.0.group_by_fields.0", "host"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.status", "high"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "case.0.condition", "a > 0"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		testCheckResourceAttrStandardRule(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
	)
}

func testAccCheckDatadogSecurityMonitoringRuleExists(accProvider func() (*schema.Provider, error), rule string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, rule := range s.RootModule().Resources {
			_, _, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, rule.Primary.ID)
			if err != nil {
				return fmt.Errorf("received an error retrieving security monitoring rule: %s", err)
			}
		}
		return nil
	}
}

func testAccCheckDatadogSecurityMonitoringRuleDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_security_monitoring_rule" {
				_, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, resource.Primary.ID)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == 404 {
						continue
					}
					return fmt.Errorf("received an error deleting security monitoring rule: %s", err)
				}
				return fmt.Errorf("monitor still exists")
			}
		}
		return nil
	}

}
