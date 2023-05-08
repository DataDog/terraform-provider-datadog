package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccTeamMembershipBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamMembershipDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamMembership(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamMembershipExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(
						"datadog_team_membership.foo", "role", "UPDATE ME"),
				),
			},
		},
	})
}

func testAccCheckDatadogTeamMembership(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_team_membership" "foo" {
    team_id = "UPDATE ME"
    role = "UPDATE ME"
}`, uniq)
}

func testAccCheckDatadogTeamMembershipDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := TeamMembershipDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func TeamMembershipDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_team_membership" {
				continue
			}
			teamId := r.Primary.Attributes["team_id"]
			pageSize := r.Primary.Attributes["page[size]"]
			pageNumber := r.Primary.Attributes["page[number]"]
			sort := r.Primary.Attributes["sort"]
			filterKeyword := r.Primary.Attributes["filter[keyword]"]

			_, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamMemberships(auth, teamId, pageSize, pageNumber, sort, filterKeyword)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving TeamMembership %s", err)}
			}
			return &utils.RetryableError{Prob: "TeamMembership still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogTeamMembershipExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := teamMembershipExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func teamMembershipExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_team_membership" {
			continue
		}
		teamId := r.Primary.Attributes["team_id"]
		pageSize := r.Primary.Attributes["page[size]"]
		pageNumber := r.Primary.Attributes["page[number]"]
		sort := r.Primary.Attributes["sort"]
		filterKeyword := r.Primary.Attributes["filter[keyword]"]

		_, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamMemberships(auth, teamId, pageSize, pageNumber, sort, filterKeyword)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving TeamMembership")
		}
	}
	return nil
}
