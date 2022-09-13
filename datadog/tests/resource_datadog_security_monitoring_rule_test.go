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

func testAccCheckDatadogSecurityMonitoringCreatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_rule" "acceptance_test" {
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
`, name)
}

func testAccCheckDatadogSecurityMonitorCreatedCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
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
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "u:tomato"),
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
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "u:tomato"),
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
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "u:tomato"),
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
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "u:tomato"),
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
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "i:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "u:tomato"),
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
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "u:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "i:tomato"),
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
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "u:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "i:tomato"),
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
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.0", "u:tomato"),
		resource.TestCheckResourceAttr(
			tfSecurityRuleName, "tags.1", "i:tomato"),
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
