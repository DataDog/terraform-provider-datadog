package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccTeamMembershipBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamMembershipDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamMembership(uniq, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamMembershipExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(
						"datadog_team_membership.foo", "role", "admin"),
				),
			},
		},
	})
}

func testAccCheckDatadogTeamMembership(uniq, username string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_team" "foo" {
	description = "Example team"
	handle      = "%s"
	name        = "%s"
}
	  
resource "datadog_user" "foo" {
	email = "%s"
}
	  
# Create new team_membership resource
resource "datadog_team_membership" "foo" {
	team_id = datadog_team.foo.id
	user_id = datadog_user.foo.id
	role    = "admin"
}`, uniq, uniq, username)
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
			teamId := r.Primary.ID
			r, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamMemberships(auth, teamId)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving TeamMembership %s", err)}
			}
			for _, team := range r.Data {
				if team.GetId() == teamId {
					return &utils.RetryableError{Prob: "TeamMembership still exists"}
				}
			}
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
		teamId := r.Primary.ID
		r, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamMemberships(auth, teamId)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving TeamMembership")
		}
		for _, team := range r.Data {
			if team.GetId() == teamId {
				return nil
			}
		}
		return utils.TranslateClientError(err, httpResp, "error retrieving TeamMembership")
	}
	return nil
}
