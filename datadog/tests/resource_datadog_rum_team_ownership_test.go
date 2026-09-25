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

func TestAccDatadogRumTeamOwnership(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	resourceName := "datadog_rum_team_ownership.browser"
	var originalID string
	var replacementID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumTeamOwnershipDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogRumTeamOwnershipConfig(uniq, "exact", "/checkout"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumTeamOwnershipExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "datadog_rum_application.app", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "team_handle", "datadog_team.owner", "handle"),
					resource.TestCheckResourceAttr(resourceName, "match_type", "exact"),
					resource.TestCheckResourceAttr(resourceName, "service", ""),
					resource.TestCheckResourceAttrSet(resourceName, "org_id"),
					resource.TestCheckResourceAttr("datadog_rum_team_ownership.mobile", "application_id", "00000000-0000-0000-0000-000000000000"),
					resource.TestCheckResourceAttr("datadog_rum_team_ownership.mobile", "service", uniq),
					resource.TestCheckResourceAttr("datadog_rum_team_ownership.mobile", "match_type", "prefix"),
					resource.TestCheckResourceAttr("datadog_rum_team_ownership.global", "application_id", "00000000-0000-0000-0000-000000000000"),
					resource.TestCheckResourceAttr("datadog_rum_team_ownership.global", "service", ""),
					resource.TestCheckResourceAttr("datadog_rum_team_ownership.global", "match_type", "exact"),
					func(s *terraform.State) error {
						originalID = s.RootModule().Resources[resourceName].Primary.ID
						return nil
					},
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDatadogRumTeamOwnershipConfig(uniq, "prefix", "/checkout/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumTeamOwnershipExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "match_type", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "view_name", "/checkout/"),
					func(s *terraform.State) error {
						replacementID = s.RootModule().Resources[resourceName].Primary.ID
						if replacementID == originalID {
							return fmt.Errorf("expected mapping replacement, ID is still %s", replacementID)
						}
						return nil
					},
				),
			},
			{
				PreConfig: func() {
					api := datadogV2.NewRumTeamsOwnershipApi(providers.frameworkProvider.DatadogApiInstances.HttpClient)
					if _, err := api.DeleteTeamsOwnershipMapping(providers.frameworkProvider.Auth, replacementID); err != nil {
						t.Fatalf("failed to delete RUM team ownership mapping outside Terraform: %s", err)
					}
				},
				Config:             testAccDatadogRumTeamOwnershipConfig(uniq, "prefix", "/checkout/"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDatadogRumTeamOwnershipConfig(uniq, matchType, viewName string) string {
	return fmt.Sprintf(`
resource "datadog_team" "owner" {
  description = "Terraform RUM ownership acceptance test."
  handle      = %[1]q
  name        = %[1]q
}

resource "datadog_rum_application" "app" {
  name = %[1]q
  type = "browser"
}

resource "datadog_rum_team_ownership" "browser" {
  application_id = datadog_rum_application.app.id
  match_type     = %[2]q
  team_handle    = datadog_team.owner.handle
  view_name      = %[3]q
}

resource "datadog_rum_team_ownership" "mobile" {
  service     = %[1]q
  match_type  = "prefix"
  team_handle = datadog_team.owner.handle
  view_name   = "Mobile"
}

resource "datadog_rum_team_ownership" "global" {
  match_type  = "exact"
  team_handle = datadog_team.owner.handle
  view_name   = "Global"
}
`, uniq, matchType, viewName)
}

func testAccCheckDatadogRumTeamOwnershipExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		api := datadogV2.NewRumTeamsOwnershipApi(accProvider.DatadogApiInstances.HttpClient)
		_, httpResp, err := api.GetTeamsOwnershipMapping(accProvider.Auth, rs.Primary.ID)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving RUM team ownership mapping")
		}
		return nil
	}
}

func testAccCheckDatadogRumTeamOwnershipDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := datadogV2.NewRumTeamsOwnershipApi(accProvider.DatadogApiInstances.HttpClient)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "datadog_rum_team_ownership" {
				continue
			}

			_, httpResp, err := api.GetTeamsOwnershipMapping(accProvider.Auth, rs.Primary.ID)
			if err != nil && httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error checking RUM team ownership mapping destruction")
			}
			return fmt.Errorf("RUM team ownership mapping %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
