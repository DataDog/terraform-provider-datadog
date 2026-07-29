package test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

//go:embed resource_datadog_on_call_team_routing_rules_test.tf
var OnCallTeamRoutingRulesTest string

func TestAccOnCallTeamRoutingRulesCreateAndUpdate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	userEmail := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	namePrefix := "team-" + uniq
	handlePrefix := "team-" + uniq

	createConfig := func(effectiveDate string) string {
		return strings.NewReplacer(
			"USER_EMAIL", userEmail,
			"POLICY_NAME", uniq,
			"UNIQ", uniq,
			"TEAM_HANDLE", handlePrefix,
			"TEAM_NAME", namePrefix,
		).Replace(OnCallTeamRoutingRulesTest)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallTeamRoutingRulesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: createConfig("2025-01-01T00:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallTeamRoutingRulesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_on_call_team_routing_rules.team_rules_test", "rule.0.query", "tags.service:test"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_team_routing_rules.team_rules_test", "rule.1.query", "tags.service:payment"),
					resource.TestCheckResourceAttrSet(
						"datadog_on_call_team_routing_rules.team_rules_test", "rule.1.action.0.escalation_policy.policy_id"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_team_routing_rules.team_rules_test", "rule.1.action.0.escalation_policy.ack_timeout_minutes", "30"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_team_routing_rules.team_rules_test", "rule.1.action.0.escalation_policy.urgency", "low"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_team_routing_rules.team_rules_test", "rule.1.action.0.escalation_policy.support_hours.restriction.#", "5"),
				),
			},
		},
	})
}

func testAccCheckDatadogOnCallTeamRoutingRulesExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_on_call_team_routing_rules" {
				continue
			}
			id := r.Primary.ID

			rules, httpResp, err := apiInstances.GetOnCallApiV2().GetOnCallTeamRoutingRules(auth, id)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving OnCallTeamRoutingRules")
			}

			if rules.Data == nil || rules.Data.Relationships.Rules == nil || len(rules.Data.Relationships.Rules.Data) == 0 {
				return errors.New("OnCallTeamRoutingRules not found")
			}

		}
		return nil
	}
}

func testAccCheckDatadogOnCallTeamRoutingRulesDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return utils.Retry(2, 10, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "resource_datadog_on_call_team_routing_rules" {
					continue
				}
				id := r.Primary.ID

				rules, _, err := apiInstances.GetOnCallApiV2().GetOnCallTeamRoutingRules(auth, id)
				if err != nil {
					if rules.Data == nil || rules.Data.Relationships.Rules == nil || len(rules.Data.Relationships.Rules.Data) == 0 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("error retrieving OnCallTeamRoutingRules %s", err)}
				}
				return &utils.RetryableError{Prob: "OnCallTeamRoutingRules still exists"}
			}
			return nil
		})
	}
}

func TestAccOnCallTeamRoutingRulesCatchAllValidation(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "datadog_on_call_team_routing_rules" "catch_all_validation" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    query             = "tags.service:test"
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				  }
				}`,
				ExpectError: regexp.MustCompile("invalid query on last rule"),
			},
			{
				Config: `
				resource "datadog_on_call_team_routing_rules" "catch_all_validation" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				    time_restrictions {
				      time_zone = "America/New_York"
				      restriction {
				        end_day    = "monday"
				        end_time   = "17:00:00"
				        start_day  = "monday"
				        start_time = "09:00:00"
				      }
				    }
				  }
				}`,
				ExpectError: regexp.MustCompile("invalid time_restrictions on last rule"),
			},
			{
				Config: `
				resource "datadog_on_call_team_routing_rules" "catch_all_validation" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    query             = "tags.service:test"
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				  }
				  rule {
				    action {
				      send_slack_message {
				        workspace = "workspace"
				        channel   = "channel"
				      }
				    }
				  }
				}`,
				ExpectError: regexp.MustCompile("missing escalation policy on last rule"),
			},
		},
	})
}

func TestAccOnCallTeamRoutingRulesUnknownBlocks(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	unknownSet := `
	resource "datadog_role" "unknown" {
	  name = "tf-test-unknown-blocks"
	}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Unknown `rule` list.
				Config: unknownSet + `
				resource "datadog_on_call_team_routing_rules" "unknown_blocks" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  dynamic "rule" {
				    for_each = toset(split(",", datadog_role.unknown.id))
				    content {
				      escalation_policy = "00000000-aba2-0000-0000-000000000001"
				    }
				  }
				}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Unknown `action` list nested in an otherwise known rule.
				Config: unknownSet + `
				resource "datadog_on_call_team_routing_rules" "unknown_blocks" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    query             = "tags.service:test"
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				  }
				  rule {
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				    dynamic "action" {
				      for_each = toset(split(",", datadog_role.unknown.id))
				      content {
				        trigger_workflow_automation {
				          handle = "handle"
				        }
				      }
				    }
				  }
				}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// An unknown scalar is not a reason to skip validation: the
				// block structure is known, so the catch-all checks still run.
				Config: unknownSet + `
				resource "datadog_on_call_team_routing_rules" "unknown_blocks" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    query             = "tags.service:test"
				    escalation_policy = datadog_role.unknown.id
				  }
				}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("invalid query on last rule"),
			},
			{
				// A catch-all violation that is already known at plan time must
				// still surface even when another rule carries an unknown block.
				// Skipping all validation on any unknown structure (the earlier
				// regression) would let this reach apply and fail there instead.
				Config: unknownSet + `
				resource "datadog_on_call_team_routing_rules" "unknown_blocks" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				    dynamic "action" {
				      for_each = toset(split(",", datadog_role.unknown.id))
				      content {
				        trigger_workflow_automation {
				          handle = "handle"
				        }
				      }
				    }
				  }
				  rule {
				    query             = "tags.service:test"
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				  }
				}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("invalid query on last rule"),
			},
		},
	})
}
