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

func TestAccTeamConnectionBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamConnectionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_team" "conn_team" {
	description = "Team for connection test"
	handle      = "%s"
	name        = "%s"
}

resource "datadog_team_connection" "foo" {
	team {
		id   = datadog_team.conn_team.id
		type = "team"
	}
	connected_team {
		id   = "@DataDog/%s"
		type = "github_team"
	}
}
`, uniq, uniq, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamConnectionExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrSet("datadog_team_connection.foo", "id"),
					resource.TestCheckResourceAttrPair("datadog_team_connection.foo", "team.id", "datadog_team.conn_team", "id"),
					resource.TestCheckResourceAttr("datadog_team_connection.foo", "team.type", "team"),
					resource.TestCheckResourceAttr("datadog_team_connection.foo", "connected_team.id", fmt.Sprintf("@DataDog/%s", uniq)),
					resource.TestCheckResourceAttr("datadog_team_connection.foo", "connected_team.type", "github_team"),
					resource.TestCheckResourceAttrSet("datadog_team_connection.foo", "source"),
				),
			},
			{
				ResourceName:      "datadog_team_connection.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTeamConnectionWithOptionalFields(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamConnectionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_team" "conn_team" {
	description = "Team for connection test"
	handle      = "%s"
	name        = "%s"
}

resource "datadog_team_connection" "foo" {
	team {
		id   = datadog_team.conn_team.id
		type = "team"
	}
	connected_team {
		id   = "@DataDog/%s"
		type = "github_team"
	}
	source     = "github"
}
`, uniq, uniq, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamConnectionExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrSet("datadog_team_connection.foo", "id"),
					resource.TestCheckResourceAttrPair("datadog_team_connection.foo", "team.id", "datadog_team.conn_team", "id"),
					resource.TestCheckResourceAttr("datadog_team_connection.foo", "team.type", "team"),
					resource.TestCheckResourceAttr("datadog_team_connection.foo", "connected_team.id", fmt.Sprintf("@DataDog/%s", uniq)),
					resource.TestCheckResourceAttr("datadog_team_connection.foo", "connected_team.type", "github_team"),
					resource.TestCheckResourceAttr("datadog_team_connection.foo", "source", "github"),
				),
			},
		},
	})
}

func testAccCheckDatadogTeamConnectionExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources["datadog_team_connection.foo"]
		if !ok {
			return fmt.Errorf("datadog_team_connection.foo not found in state")
		}

		id := r.Primary.ID
		opts := datadogV2.NewListTeamConnectionsOptionalParameters().WithFilterConnectionIds([]string{id})
		resp, httpResp, err := accProvider.DatadogApiInstances.GetTeamsApiV2().ListTeamConnections(accProvider.Auth, *opts)
		if err != nil {
			return fmt.Errorf("error retrieving TeamConnection %s: %s", id, utils.TranslateClientError(err, httpResp, ""))
		}

		if len(resp.GetData()) == 0 {
			return fmt.Errorf("TeamConnection %s not found", id)
		}
		return nil
	}
}

func testAccCheckDatadogTeamConnectionDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_team_connection" {
				continue
			}
			id := r.Primary.ID

			opts := datadogV2.NewListTeamConnectionsOptionalParameters().WithFilterConnectionIds([]string{id})
			resp, httpResp, err := accProvider.DatadogApiInstances.GetTeamsApiV2().ListTeamConnections(accProvider.Auth, *opts)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("error retrieving TeamConnection %s: %s", id, err)
			}

			if len(resp.GetData()) > 0 {
				return fmt.Errorf("TeamConnection %s still exists", id)
			}
		}
		return nil
	}
}
