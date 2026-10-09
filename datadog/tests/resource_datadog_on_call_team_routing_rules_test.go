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

//go:embed resource_datadog_on_call_team_routing_rules_reroute_test.tf
var OnCallTeamRoutingRulesRerouteTest string

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

func TestAccOnCallTeamRoutingRulesRerouteToTeam(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	createConfig := func(firstDest, secondDest string) string {
		return strings.NewReplacer(
			"TEAM_HANDLE", "team-"+uniq,
			"TEAM_NAME", "team-"+uniq,
			"FIRST_DEST", firstDest,
			"SECOND_DEST", secondDest,
		).Replace(OnCallTeamRoutingRulesRerouteTest)
	}

	resourceName := "datadog_on_call_team_routing_rules.reroute_test"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallTeamRoutingRulesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: createConfig("reroute_dest_a", "reroute_dest_b"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallTeamRoutingRulesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.query", "priority:2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.action.0.reroute_to_team.destination_team_id", "datadog_team.reroute_dest_a", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule.1.action.0.reroute_to_team.destination_team_id", "datadog_team.reroute_dest_b", "id"),
				),
			},
			{
				Config: createConfig("reroute_dest_b", "reroute_dest_a"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallTeamRoutingRulesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrPair(resourceName, "rule.0.action.0.reroute_to_team.destination_team_id", "datadog_team.reroute_dest_b", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "rule.1.action.0.reroute_to_team.destination_team_id", "datadog_team.reroute_dest_a", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOnCallTeamRoutingRulesRerouteToTeamValidation(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "datadog_on_call_team_routing_rules" "reroute_validation" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    action {
				      reroute_to_team {
				        destination_team_id = "00000000-aba2-0000-0000-000000000002"
				      }
				    }
				    action {
				      escalation_policy {
				        policy_id = "00000000-aba2-0000-0000-000000000001"
				      }
				    }
				  }
				}`,
				ExpectError: regexp.MustCompile("incompatible action with reroute_to_team"),
			},
			{
				Config: `
				resource "datadog_on_call_team_routing_rules" "reroute_validation" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				    action {
				      reroute_to_team {
				        destination_team_id = "00000000-aba2-0000-0000-000000000002"
				      }
				    }
				  }
				}`,
				ExpectError: regexp.MustCompile("incompatible action with reroute_to_team"),
			},
			{
				Config: `
				resource "datadog_on_call_team_routing_rules" "reroute_validation" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    action {
				      reroute_to_team {
				        destination_team_id = "00000000-aba2-0000-0000-000000000002"
				      }
				    }
				    action {
				      reroute_to_team {
				        destination_team_id = "00000000-aba2-0000-0000-000000000003"
				      }
				    }
				  }
				}`,
				ExpectError: regexp.MustCompile("duplicate reroute_to_team action"),
			},
			{
				Config: `
				resource "datadog_on_call_team_routing_rules" "reroute_validation" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    action {
				      reroute_to_team {}
				    }
				  }
				}`,
				ExpectError: regexp.MustCompile("missing destination_team_id"),
			},
			{
				Config: `
				resource "datadog_on_call_team_routing_rules" "reroute_validation" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    action {
				      reroute_to_team {
				        destination_team_id = "not-a-uuid"
				      }
				    }
				  }
				}`,
				ExpectError: regexp.MustCompile(`must be a\s+valid UUID`),
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
				    query = "tags.service:test"
				  }
				  rule {
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				  }
				}`,
				ExpectError: regexp.MustCompile("missing actions on rule"),
			},
		},
	})
}

// TestAccOnCallTeamRoutingRulesUnknownBlocks guards against the crash fixed
// here and previously in #3862: `rule`/`action`/`restriction` are plain Go
// slices, which can't hold a not-yet-known value. Decoding an unresolved
// `dynamic` block straight into the typed model (as ValidateConfig used to)
// crashes with "Received unknown value... Suggested Type: basetypes.ListValue".
// There is intentionally no ValidateConfig on this resource: it can't
// validate a config it can't fully decode without reintroducing that crash,
// so all of `onCallTeamRoutingRulesModel.Validate` runs only in Create/Update,
// once every value is known. These steps just assert `plan` never crashes,
// regardless of where the unknown collection sits.
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
				// An unknown scalar (as opposed to an unknown collection)
				// never crashed even before this fix: types.String natively
				// supports being unknown.
				Config: unknownSet + `
				resource "datadog_on_call_team_routing_rules" "unknown_blocks" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    query             = "tags.service:test"
				    escalation_policy = datadog_role.unknown.id
				  }
				}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// A rule with both a KNOWN, already-invalid `query` (it's
				// also the last/catch-all rule) and an unrelated unknown
				// `action` list. There's no plan-time validation at all now,
				// so this just asserts the plan doesn't crash; the invalid
				// query is still caught later, at Create/Update.
				Config: unknownSet + `
				resource "datadog_on_call_team_routing_rules" "unknown_blocks" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    query             = "tags.service:test"
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
				// Unknown `restriction` list nested two levels deep, under
				// `rule.time_restrictions`. Same underlying bug as an
				// unknown `rule`/`action` list: teamTimeRestrictionsModel.Restrictions
				// is also a plain []*restrictionsModel slice.
				Config: unknownSet + `
				resource "datadog_on_call_team_routing_rules" "unknown_blocks" {
				  id = "00000000-aba2-0000-0000-000000000000"
				  rule {
				    escalation_policy = "00000000-aba2-0000-0000-000000000001"
				    time_restrictions {
				      time_zone = "UTC"
				      dynamic "restriction" {
				        for_each = toset(split(",", datadog_role.unknown.id))
				        content {
				          start_day  = "Monday"
				          start_time = "09:00:00"
				          end_day    = "Monday"
				          end_time   = "17:00:00"
				        }
				      }
				    }
				  }
				}`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
