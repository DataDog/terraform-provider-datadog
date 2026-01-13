package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogTeamNotificationRuleDatasourceBasic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTeamNotificationRuleConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					// Check data source that fetches all rules for team foo
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo_all", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo_all", "team_id"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo_all", "notification_rules.#", "1"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo_all", "notification_rules.0.id"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo_all", "notification_rules.0.email.enabled", "true"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo_all", "notification_rules.0.ms_teams.connector_name", "test-teams-handle"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo_all", "notification_rules.0.pagerduty.service_name", "my-service"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo_all", "notification_rules.0.slack.channel", "#test-channel"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo_all", "notification_rules.0.slack.workspace", "Datadog"),
					// Check data source that fetches specific rule for team foo
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo_specific", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo_specific", "team_id"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo_specific", "rule_id"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo_specific", "notification_rules.#", "1"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo_specific", "notification_rules.0.id"),
					// Check data source that fetches all rules for team foo2
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo2_all", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo2_all", "team_id"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo2_all", "notification_rules.#", "1"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.foo2_all", "notification_rules.0.id"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo2_all", "notification_rules.0.email.enabled", "false"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo2_all", "notification_rules.0.ms_teams.connector_name", "test-teams-handle-2"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo2_all", "notification_rules.0.pagerduty.service_name", "my-service-2"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo2_all", "notification_rules.0.slack.channel", "#test-channel-2"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.foo2_all", "notification_rules.0.slack.workspace", "Datadog2"),
				),
			},
		},
	})
}

func TestAccDatadogTeamNotificationRuleDatasourceEmpty(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTeamNotificationRuleEmptyConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					// Check that team with no notification rules returns empty list
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.empty", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rules.empty", "team_id"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rules.empty", "notification_rules.#", "0"),
				),
			},
		},
	})
}

func testAccDatasourceTeamNotificationRuleConfig(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_team" "foo" {
			description = "Example team"
			handle      = "%s"
			name        = "%s"
		}

		resource "datadog_team" "foo2" {
			description = "Example team 2"
			handle      = "%s-2"
			name        = "%s-2"
		}

		resource "datadog_team_notification_rule" "foo" {
		  team_id = datadog_team.foo.id
		  email {
			enabled = true
		  }
		  ms_teams {
			connector_name = "test-teams-handle"
		  }
		  pagerduty {
			service_name = "my-service"
		  }
		  slack {
			channel   = "#test-channel"
			workspace = "Datadog"
		  }
		}

		resource "datadog_team_notification_rule" "foo2" {
		  team_id = datadog_team.foo2.id
		  email {
			enabled = false
		  }
		  ms_teams {
			connector_name = "test-teams-handle-2"
		  }
		  pagerduty {
			service_name = "my-service-2"
		  }
		  slack {
			channel   = "#test-channel-2"
			workspace = "Datadog2"
		  }
		}

		# Data source that fetches all rules for team foo
		data "datadog_team_notification_rules" "foo_all" {
			team_id    = datadog_team.foo.id
			depends_on = [datadog_team_notification_rule.foo]
		}

		# Data source that fetches a specific rule for team foo
		data "datadog_team_notification_rules" "foo_specific" {
			team_id    = datadog_team.foo.id
			rule_id    = datadog_team_notification_rule.foo.id
			depends_on = [datadog_team_notification_rule.foo]
		}

		# Data source that fetches all rules for team foo2
		data "datadog_team_notification_rules" "foo2_all" {
			team_id    = datadog_team.foo2.id
			depends_on = [datadog_team_notification_rule.foo2]
		}
`, uniq, uniq, uniq, uniq)
}

func testAccDatasourceTeamNotificationRuleEmptyConfig(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_team" "empty" {
			description = "Team with no notification rules"
			handle      = "%s-empty"
			name        = "%s-empty"
		}

		# Data source that fetches all rules for team with no rules
		data "datadog_team_notification_rules" "empty" {
			team_id = datadog_team.empty.id
		}
`, uniq, uniq)
}
