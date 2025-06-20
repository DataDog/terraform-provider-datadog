package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccScorecardRuleBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqName := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogScorecardRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccScorecardRule(uniqName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogScorecardRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "name", uniqName),
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "owner", "terraform"),
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "scorecard_name", "terraform_scorecard"),
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "enabled", "true"),
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "custom", "true"),
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "scope_query", "env:prod team:platform"),
					resource.TestCheckResourceAttrSet("datadog_service_scorecard_rule.foo", "created_at"),
					resource.TestCheckResourceAttrSet("datadog_service_scorecard_rule.foo", "modified_at"),
				),
			},
			{
				Config: testAccScorecardRuleUpdated(uniqName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "description", "Updated description"),
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "level", "1"),
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "custom", "true"),
					resource.TestCheckResourceAttr("datadog_service_scorecard_rule.foo", "scope_query", "env:staging team:development"),
					resource.TestCheckResourceAttrSet("datadog_service_scorecard_rule.foo", "created_at"),
					resource.TestCheckResourceAttrSet("datadog_service_scorecard_rule.foo", "modified_at"),
				),
			},
		},
	})
}

func testAccScorecardRule(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_scorecard_rule" "foo" {
  name           = "%s"
  description    = "Acceptance test rule"
  enabled        = true
  owner          = "terraform"
  scorecard_name = "terraform_scorecard"
  level          = 2
  scope_query    = "env:prod team:platform"
}
`, uniq)
}

func testAccScorecardRuleUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_scorecard_rule" "foo" {
  name           = "%s"
  description    = "Updated description"
  enabled        = true
  owner          = "terraform"
  scorecard_name = "terraform_scorecard"
  level          = 1
  scope_query    = "env:staging team:development"
}
`, uniq)
}

func testAccCheckDatadogScorecardRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_scorecard_rule" {
				continue
			}
			id := r.Primary.ID
			opt := datadogV2.NewListScorecardRulesOptionalParameters().WithFilterRuleId(id)
			resp, httpResp, err := apiInstances.GetServiceScorecardsApiV2().ListScorecardRules(auth, *opt)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving Scorecard Rule")
			}
			if len(resp.Data) == 0 {
				return fmt.Errorf("scorecard rule not found")
			}
		}
		return nil
	}
}

func testAccCheckDatadogScorecardRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_scorecard_rule" {
				continue
			}
			id := r.Primary.ID

			err := utils.Retry(2, 10, func() error {
				opt := datadogV2.NewListScorecardRulesOptionalParameters().WithFilterRuleId(id)
				resp, httpResp, err := apiInstances.GetServiceScorecardsApiV2().ListScorecardRules(auth, *opt)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("error retrieving Scorecard Rule: %s", err)}
				}
				if len(resp.Data) == 0 {
					return nil
				}
				return &utils.RetryableError{Prob: "Scorecard Rule still exists"}
			})

			if err != nil {
				return err
			}
		}
		return nil
	}
}
