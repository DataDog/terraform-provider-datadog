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

func TestAccTeamLinkBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamLinkDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamLink(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamLinkExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(
						"datadog_team_link.foo", "label", "Link label"),
					resource.TestCheckResourceAttr(
						"datadog_team_link.foo", "position", "1"),
					resource.TestCheckResourceAttr(
						"datadog_team_link.foo", "url", "https://example.com"),
				),
			},
			{
				Config: testAccCheckDatadogTeamLinkUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamLinkExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(
						"datadog_team_link.foo", "label", "Link label updated"),
					resource.TestCheckResourceAttr(
						"datadog_team_link.foo", "position", "2"),
					resource.TestCheckResourceAttr(
						"datadog_team_link.foo", "url", "https://example.com"),
					resource.TestCheckResourceAttr(
						"datadog_team_link.bar", "label", "Link label"),
					resource.TestCheckResourceAttr(
						"datadog_team_link.bar", "position", "1"),
					resource.TestCheckResourceAttr(
						"datadog_team_link.bar", "url", "https://example.com"),
				),
			},
		},
	})
}

func testAccCheckDatadogTeamLink(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_team" "foo" {
	description = "123"
	handle      = "%s"
	name        = "%s"
}

resource "datadog_team_link" "foo" {
    label = "Link label"
    position = 1
    team_id = datadog_team.foo.id
    url = "https://example.com"
}`, uniq, uniq)
}

func testAccCheckDatadogTeamLinkUpdated(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_team" "foo" {
	description = "123"
	handle      = "%s"
	name        = "%s"
}

resource "datadog_team_link" "foo" {
    label = "Link label updated"
    position = 2
    team_id = datadog_team.foo.id
    url = "https://example.com"
}

resource "datadog_team_link" "bar" {
    label = "Link label"
    position = 1
    team_id = datadog_team.foo.id
    url = "https://example.com"
}
`, uniq, uniq)
}

func testAccCheckDatadogTeamLinkDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := TeamLinkDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func TeamLinkDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_team_link" {
				continue
			}
			teamId := r.Primary.Attributes["team_id"]
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamLink(auth, teamId, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving TeamLink %s", err)}
			}
			return &utils.RetryableError{Prob: "TeamLink still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogTeamLinkExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := teamLinkExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func teamLinkExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_team_link" {
			continue
		}
		teamId := r.Primary.Attributes["team_id"]
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamLink(auth, teamId, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving TeamLink")
		}
	}
	return nil
}
