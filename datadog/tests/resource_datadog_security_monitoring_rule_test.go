package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

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

func TestAccDatadogSecurityMonitoringRule_InvalidTypes(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)

	invalidValueRegex, _ := regexp.Compile("Invalid enum value")
	invalidTypeRegex, _ := regexp.Compile("Incorrect attribute value type")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogSecurityMonitoringRule("\"infrastructure_configuration\"", ruleName),
				ExpectError: invalidValueRegex,
			},
			{
				Config:      testAccCheckDatadogSecurityMonitoringRule("\"cloud_configuration\"", ruleName),
				ExpectError: invalidValueRegex,
			},
			{
				Config:      testAccCheckDatadogSecurityMonitoringRule("\"bogus_type\"", ruleName),
				ExpectError: invalidValueRegex,
			},
			{
				Config:      testAccCheckDatadogSecurityMonitoringRule("[\"one\", \"two\"]", ruleName),
				ExpectError: invalidTypeRegex,
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
`, suffix, name)
}

func testAccCheckDatadogSecurityMonitorCreatedCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return testAccCheckDatadogSecurityMonitorCreatedCheckWithId(accProvider, ruleName, "")
}

func testAccCheckDatadogSecurityMonitorCreatedCheckWithId(accProvider func() (*schema.Provider, error), ruleName string, id string) resource.TestCheckFunc {
	tfSecurityRuleNameWithId := tfSecurityRuleName
	if id != "" {
		tfSecurityRuleNameWithId = fmt.Sprintf("%s_%s", tfSecurityRuleNameWithId, id)
	}
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "has_extended_title", "true"),
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
			tfSecurityRuleName, "query.2.name", "third"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.2.query", "does not really match much either"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.2.aggregation", "sum"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.2.group_by_fields.0", "host"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.2.metric", "@network.bytes_read"),
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
			tfSecurityRuleName, "case.2.name", "low case"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.2.status", "low"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.2.condition", "third > 9000"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.decrease_criticality_based_on_env", "true"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.0.query", "does not really suppress"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.1.action", "require"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.1.query", "does not really require neither"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "i:tomato"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "u:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringCreatedConfigNewValueRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
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
`, name)
}

func testAccCheckDatadogSecurityMonitorCreatedCheckNewValueRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
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
			tfSecurityRuleName, "query.0.aggregation", "new_value"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "host"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", ""),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "high"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.detection_method", "new_value"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.new_value_options.0.forget_after", "7"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_method", "duration"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_duration", "1"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_threshold", "0"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "i:tomato"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "u:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringCreatedConfigImpossibleTravelRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
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
`, name)
}

func testAccCheckDatadogSecurityMonitorCreatedCheckImpossibleTravelRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "impossible travel rule triggered"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.name", "my_query"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.query", "*"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.aggregation", "geo_data"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "@usr.handle"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", ""),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "high"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.detection_method", "impossible_travel"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.impossible_travel_options.0.baseline_user_locations", "true"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "i:tomato"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "u:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedConfigImpossibleTravelRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
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
`, name)
}

func testAccCheckDatadogSecurityMonitorUpdatedCheckImpossibleTravelRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "impossible travel rule triggered (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.name", "my_updated_query"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.query", "*"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.aggregation", "geo_data"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "@usr.handle"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", "new case name (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "high"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.detection_method", "impossible_travel"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.impossible_travel_options.0.baseline_user_locations", "true"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "i:tomato"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "u:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringCreatedConfigCwsRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
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
`, name, strings.Replace(name, "-", "_", -1))
}

func testAccCheckDatadogSecurityMonitoringRule(ruleType string, name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
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

	type = %s
}
`, name, strings.Replace(name, "-", "_", -1), ruleType)
}

func testAccCheckDatadogSecurityMonitoringCreatedCheckCwsRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
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
			tfSecurityRuleName, "query.0.query", fmt.Sprintf("@agent.rule_id:(%s_random_id OR random_id)", strings.Replace(ruleName, "-", "_", -1))),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.aggregation", "count"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "host"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", "high case"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "high"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.condition", "first > 3"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.detection_method", "threshold"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "i:tomato"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "u:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "type", "workload_security"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
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
`, name)
}

func testAccCheckDatadogSecurityMonitoringUpdateCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName+" - updated"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "true"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "has_extended_title", "false"),
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
			tfSecurityRuleName, "options.0.decrease_criticality_based_on_env", "false"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.0.query", "does not really suppress (updated)"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "u:tomato"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "i:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedConfigNewValueRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
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
`, name)
}

func testAccCheckDatadogSecurityMonitoringUpdateCheckNewValueRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName+" - updated"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "true"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.name", "first"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.query", "does not really match much (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.aggregation", "new_value"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "service"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.metric", "@network.bytes_read"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "medium"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.detection_method", "new_value"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.new_value_options.0.forget_after", "1"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_method", "duration"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_duration", "0"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.new_value_options.0.learning_threshold", "0"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "u:tomato"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "i:tomato"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedConfigCwsRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
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
`, name, strings.Replace(name, "-", "_", -1))
}

func testAccCheckDatadogSecurityMonitoringUpdateCheckCwsRule(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "true"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.name", "first"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.query", fmt.Sprintf("@agent.rule_id:(%s_random_id OR random_id)", strings.Replace(ruleName, "-", "_", -1))),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.aggregation", "count"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "query.0.group_by_fields.0", "service"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "medium"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.condition", "first > 10"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.detection_method", "threshold"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "u:tomato"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "i:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "type", "workload_security"),
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
`, name)
}

func testAccCheckDatadogSecurityMonitoringCreatedSignalCorrelationConfig(name string) string {
	logDetectionRule0 := fmt.Sprintf("%s_rule_0", name)
	logDetectionRule1 := fmt.Sprintf("%s_rule_1", name)
	return fmt.Sprintf(`
%s
%s
resource "datadog_security_monitoring_rule" "acceptance_test" {
	name = "%s"
	message = "acceptance rule triggered"
	enabled = false
	has_extended_title = true

	signal_query {
		name = "first"
		rule_id = "${datadog_security_monitoring_rule.acceptance_test_0.id}"
		aggregation = "event_count"
		correlated_by_fields = ["host"]
	}

	signal_query {
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
`, testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(logDetectionRule0, "0"),
		testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(logDetectionRule1, "1"),
		name)
}

func testCheckResourceAttrPairSignalCorrelation(name string, queryId int, logDetectionId int) resource.TestCheckFunc {
	logDetectionName := fmt.Sprintf("%s_%d", name, logDetectionId)
	signalQuery := fmt.Sprintf("signal_query.%d.rule_id", queryId)
	return resource.TestCheckResourceAttrPair(name, signalQuery, logDetectionName, "id")
}

func testAccCheckDatadogSecurityMonitorCreatedSignalCorrelationCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "has_extended_title", "true"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.0.name", "first"),
		testCheckResourceAttrPairSignalCorrelation(tfSecurityRuleName, 0, 0),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.0.aggregation", "event_count"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.0.correlated_by_fields.0", "host"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.0.correlated_query_index", ""),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.1.name", "second"),
		testCheckResourceAttrPairSignalCorrelation(tfSecurityRuleName, 1, 1),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.1.aggregation", "event_count"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.1.correlated_by_fields.0", "host"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.1.correlated_query_index", "1"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", "high case"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "high"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.condition", "first > 0 && second > 0"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.evaluation_window", "300"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "900"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.0.query", "does not really suppress"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.1.action", "require"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.1.query", "does not really require neither"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "alert:red"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "attack:advanced"),
	)
}

func testAccCheckDatadogSecurityMonitoringUpdatedSignalCorrelationConfig(name string) string {
	logDetectionRule0 := fmt.Sprintf("%s_rule_0", name)
	logDetectionRule1 := fmt.Sprintf("%s_rule_1", name)
	return fmt.Sprintf(`
%s
%s
resource "datadog_security_monitoring_rule" "acceptance_test" {
	name = "%s - updated"
	message = "acceptance rule triggered (updated)"
	enabled = true
	has_extended_title = false

	signal_query {
		name = "first_updated"
		rule_id = "${datadog_security_monitoring_rule.acceptance_test_0.id}"
		correlated_by_fields = ["service"]
		aggregation = "event_count"
	}

	signal_query {
		name = "second_updated"
		rule_id = "${datadog_security_monitoring_rule.acceptance_test_1.id}"
		correlated_by_fields = ["service"]
		correlated_query_index = "0"
		aggregation = "event_count"
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

	type = "signal_correlation"

	tags = ["alert:red", "attack:advanced"]
}
`, testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(logDetectionRule0, "0"),
		testAccCheckDatadogSecurityMonitoringCreatedConfigWithId(logDetectionRule1, "1"),
		name)
}

func testAccCheckDatadogSecurityMonitoringUpdateSignalCorrelationCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringRuleExists(accProvider, tfSecurityRuleName),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "name", ruleName+" - updated"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "message", "acceptance rule triggered (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "enabled", "true"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "has_extended_title", "false"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.0.name", "first_updated"),
		testCheckResourceAttrPairSignalCorrelation(tfSecurityRuleName, 0, 0),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.0.aggregation", "event_count"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.0.correlated_by_fields.0", "service"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.0.correlated_query_index", ""),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.1.name", "second_updated"),
		testCheckResourceAttrPairSignalCorrelation(tfSecurityRuleName, 1, 1),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.1.aggregation", "event_count"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.1.correlated_by_fields.0", "service"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "signal_query.1.correlated_query_index", "0"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.name", "high case (updated)"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.status", "medium"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.condition", "first_updated > 0 && second_updated > 0"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "case.0.notifications.0", "@user"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.evaluation_window", "60"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.keep_alive", "300"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "options.0.max_signal_duration", "600"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.0.query", "does not really suppress (updated)"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "alert:red"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "attack:advanced"),
	)
}

func testAccCheckDatadogSecurityMonitoringEnabledDefaultCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
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
			tfSecurityRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "filter.0.query", "does not really suppress (updated)"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "u:tomato"),
		resource.TestCheckTypeSetElemAttr(
			tfSecurityRuleName, "tags.*", "i:tomato"),
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

func testAccCheckDatadogSecurityMonitorCreatedRequiredCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
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
