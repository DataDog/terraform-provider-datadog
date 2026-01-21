package test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
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
