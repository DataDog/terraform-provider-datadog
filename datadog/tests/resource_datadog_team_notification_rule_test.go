package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccTeamNotificationRuleBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamNotificationRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamNotificationRuleExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func TestAccTeamNotificationRuleEmpty(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamNotificationRuleEmpty(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamNotificationRuleExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogTeamNotificationRule(uniq string) string {
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
`, uniq, uniq)
}

func testAccCheckDatadogTeamNotificationRuleEmpty(uniq string) string {
	return fmt.Sprintf(`

			resource "datadog_team" "foo" {
				description = "Example team"
				handle      = "%s"
				name        = "%s"
			}
			resource "datadog_team_notification_rule" "foo" {
			  team_id = datadog_team.foo.id
			  email {
				enabled = false
			  }
			}
`, uniq, uniq)
}

func testAccCheckDatadogTeamNotificationRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := TeamNotificationRuleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func TeamNotificationRuleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_team_notification_rule" {
				continue
			}
			teamId := r.Primary.Attributes["team_id"]
			ruleId := r.Primary.Attributes["id"]

			_, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamNotificationRule(auth, teamId, ruleId)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving TeamNotificationRule %s", err)}
			}
			return &utils.RetryableError{Prob: "TeamNotificationRule still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogTeamNotificationRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := teamNotificationRuleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func teamNotificationRuleExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_team_notification_rule" {
			continue
		}
		teamId := r.Primary.Attributes["team_id"]
		ruleId := r.Primary.Attributes["id"]

		resp, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamNotificationRule(auth, teamId, ruleId)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving TeamNotificationRule")
		}

		if httpResp.StatusCode != 200 {
			return fmt.Errorf("TeamNotificationRule %s does not exist", ruleId)
		}

		// Validate the response contains the expected ID
		if resp.Data != nil {
			if resp.Data.GetId() != ruleId {
				return fmt.Errorf("TeamNotificationRule %s does not match expected response", ruleId)
			}
		} else if resp.UnparsedObject != nil {
			// Handle unparsed case
			if dataRaw, ok := resp.UnparsedObject["data"].(map[string]interface{}); ok {
				if idVal, _ := dataRaw["id"].(string); idVal != ruleId {
					return fmt.Errorf("TeamNotificationRule %s does not match expected response", ruleId)
				}
			}
		}
	}
	return nil
}

func TestAccTeamNotificationRuleNoNotifications(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogTeamNotificationRuleNoNotifications(uniq),
				ExpectError: regexp.MustCompile("Missing Notification Configuration"),
			},
		},
	})
}

func TestAccTeamNotificationRuleImport(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamNotificationRule(uniq),
			},
			{
				ResourceName:      "datadog_team_notification_rule.foo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccTeamNotificationRuleImportStateIdFunc("datadog_team_notification_rule.foo"),
			},
		},
	})
}

func TestAccTeamNotificationRuleUpdate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamNotificationRuleEmailOnly(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "email.enabled", "true"),
					resource.TestCheckNoResourceAttr("datadog_team_notification_rule.foo", "ms_teams.connector_name"),
				),
			},
			{
				Config: testAccCheckDatadogTeamNotificationRuleSlackOnly(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "slack.channel", "#alerts"),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "slack.workspace", "MyWorkspace"),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "email.enabled", "false"),
				),
			},
			{
				Config: testAccCheckDatadogTeamNotificationRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "email.enabled", "true"),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "ms_teams.connector_name", "test-teams-handle"),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "pagerduty.service_name", "my-service"),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "slack.channel", "#test-channel"),
				),
			},
		},
	})
}

func TestAccTeamNotificationRuleRemoveNotificationType(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamNotificationRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "email.enabled", "true"),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "ms_teams.connector_name", "test-teams-handle"),
				),
			},
			{
				Config: testAccCheckDatadogTeamNotificationRuleEmailOnly(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_team_notification_rule.foo", "email.enabled", "true"),
					resource.TestCheckNoResourceAttr("datadog_team_notification_rule.foo", "ms_teams.connector_name"),
					resource.TestCheckNoResourceAttr("datadog_team_notification_rule.foo", "pagerduty.service_name"),
					resource.TestCheckNoResourceAttr("datadog_team_notification_rule.foo", "slack.channel"),
				),
			},
		},
	})
}

func testAccTeamNotificationRuleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		teamId := rs.Primary.Attributes["team_id"]
		ruleId := rs.Primary.ID

		return fmt.Sprintf("%s:%s", teamId, ruleId), nil
	}
}

func testAccCheckDatadogTeamNotificationRuleEmailOnly(uniq string) string {
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
		}
`, uniq, uniq)
}

func testAccCheckDatadogTeamNotificationRuleSlackOnly(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_team" "foo" {
			description = "Example team"
			handle      = "%s"
			name        = "%s"
		}
		resource "datadog_team_notification_rule" "foo" {
		  team_id = datadog_team.foo.id
		  email {
			enabled = false
		  }
		  slack {
			channel   = "#alerts"
			workspace = "MyWorkspace"
		  }
		}
`, uniq, uniq)
}

func testAccCheckDatadogTeamNotificationRuleNoNotifications(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_team" "foo" {
			description = "Example team"
			handle      = "%s"
			name        = "%s"
		}
		resource "datadog_team_notification_rule" "foo" {
		  team_id = datadog_team.foo.id
		}
`, uniq, uniq)
}
