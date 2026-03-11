package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogScorecardRule_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogScorecardRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogScorecardRuleConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogScorecardRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "name", uniq),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "scorecard_name", uniq+"-scorecard"),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "description", "Test rule description"),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "enabled", "true"),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "level", "1"),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "owner", "test-team"),
					resource.TestCheckResourceAttrSet("datadog_scorecard_rule.foo", "id"),
				),
			},
			{
				Config: testAccDatadogScorecardRuleConfigUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogScorecardRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "name", uniq),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "description", "Updated description"),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "enabled", "false"),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "level", "2"),
				),
			},
			{
				Config: testAccDatadogScorecardRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogScorecardRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "name", uniq),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "level", "1"),
					resource.TestCheckResourceAttr("datadog_scorecard_rule.foo", "enabled", "true"),
					resource.TestCheckNoResourceAttr("datadog_scorecard_rule.foo", "description"),
					resource.TestCheckNoResourceAttr("datadog_scorecard_rule.foo", "owner"),
				),
			},
		},
	})
}

func TestAccDatadogScorecardRule_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogScorecardRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogScorecardRuleConfig(uniq),
			},
			{
				ResourceName:      "datadog_scorecard_rule.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDatadogScorecardRuleConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_scorecard_rule" "foo" {
  name           = "%[1]s"
  scorecard_name = "%[1]s-scorecard"
  description    = "Test rule description"
  enabled        = true
  level          = "1"
  owner          = "test-team"
}
`, uniq)
}

func testAccDatadogScorecardRuleConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_scorecard_rule" "foo" {
  name           = "%[1]s"
  scorecard_name = "%[1]s-scorecard"
  description    = "Updated description"
  enabled        = false
  level          = "2"
  owner          = "test-team"
}
`, uniq)
}

func testAccDatadogScorecardRuleConfigMinimal(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_scorecard_rule" "foo" {
  name           = "%[1]s"
  scorecard_name = "%[1]s-scorecard"
  level          = "1"
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

			optParams := datadogV2.NewListScorecardRulesOptionalParameters()
			optParams.WithFilterRuleId(id)
			optParams.WithPageSize(1)
			resp, _, err := apiInstances.GetServiceScorecardsApiV2().ListScorecardRules(auth, *optParams)
			if err != nil {
				return utils.TranslateClientError(err, nil, "error checking scorecard rule exists")
			}
			if len(resp.GetData()) == 0 {
				return fmt.Errorf("scorecard rule %s not found", id)
			}
		}
		return nil
	}
}

func testAccCheckDatadogScorecardRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := scorecardRuleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func scorecardRuleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_scorecard_rule" {
				continue
			}
			id := r.Primary.ID

			optParams := datadogV2.NewListScorecardRulesOptionalParameters()
			optParams.WithFilterRuleId(id)
			optParams.WithPageSize(1)
			resp, _, err := apiInstances.GetServiceScorecardsApiV2().ListScorecardRules(auth, *optParams)
			if err != nil {
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving scorecard rule: %s", err)}
			}
			if len(resp.GetData()) > 0 {
				return &utils.RetryableError{Prob: "scorecard rule still exists"}
			}
		}
		return nil
	})
	return err
}
