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

func TestAccTeamBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeam(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_team.foo", "description", "Team description."),
					resource.TestCheckResourceAttr(
						"datadog_team.foo", "handle", uniq),
					resource.TestCheckResourceAttr(
						"datadog_team.foo", "name", uniq),
				),
			},
			{
				Config: testAccCheckDatadogTeamUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_team.foo", "description", "Team description - updated."),
					resource.TestCheckResourceAttr(
						"datadog_team.foo", "handle", fmt.Sprintf("%s-updated", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_team.foo", "name", fmt.Sprintf("%s-updated", uniq)),
				),
			},
		},
	})
}

func testAccCheckDatadogTeam(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_team" "foo" {
    description = "Team description."
    handle = "%s"
    name = "%s"
}`, uniq, uniq)
}

func testAccCheckDatadogTeamUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_team" "foo" {
    description = "Team description - updated."
    handle = "%s-updated"
    name = "%s-updated"
}`, uniq, uniq)
}

func testAccCheckDatadogTeamDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := TeamDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func TeamDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_team" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetTeamsApiV2().GetTeam(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Team %s", err)}
			}
			return &utils.RetryableError{Prob: "Team still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogTeamExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := teamExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func teamExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_team" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetTeamsApiV2().GetTeam(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Team")
		}
	}
	return nil
}
