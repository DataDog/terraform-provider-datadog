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
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTeamNotificationRuleSingularConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					// Check data source that fetches specific rule for team foo
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rule.foo_specific", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rule.foo_specific", "team_id"),
					resource.TestCheckResourceAttrSet("data.datadog_team_notification_rule.foo_specific", "rule_id"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rule.foo_specific", "email.enabled", "true"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rule.foo_specific", "ms_teams.connector_name", "test-teams-handle"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rule.foo_specific", "pagerduty.service_name", "my-service"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rule.foo_specific", "slack.channel", "#test-channel"),
					resource.TestCheckResourceAttr("data.datadog_team_notification_rule.foo_specific", "slack.workspace", "Datadog"),
				),
			},
		},
	})
}

func testAccDatasourceTeamNotificationRuleSingularConfig(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_team" "foo" {
			description = "Example team"
			handle      = "%s"
			name        = "%s"
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

		# Data source that fetches a specific rule for team foo
		data "datadog_team_notification_rule" "foo_specific" {
			team_id    = datadog_team.foo.id
			rule_id    = datadog_team_notification_rule.foo.id
			depends_on = [datadog_team_notification_rule.foo]
		}
`, uniq, uniq)
}
