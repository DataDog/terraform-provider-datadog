package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccTeamHierarchyLinksBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTeamHierarchyLinksDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamHierarchyLinks(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTeamHierarchyLinksExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogTeamHierarchyLinks(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_team-hierarchy-links" "foo" {
    body {
    data {
    relationships {
    parent_team {
    data {
    id = "692e8073-12c4-4c71-8408-5090bd44c9c8"
    type = "team"
    }
    }
    sub_team {
    data {
    id = "692e8073-12c4-4c71-8408-5090bd44c9c8"
    type = "team"
    }
    }
    }
    type = "team_hierarchy_links"
    }
    }
}`, uniq)
}

func testAccCheckDatadogTeamHierarchyLinksDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := TeamHierarchyLinksDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func TeamHierarchyLinksDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_team-hierarchy-links" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamHierarchyLink(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving TeamHierarchyLinks %s", err)}
			}
			return &utils.RetryableError{Prob: "TeamHierarchyLinks still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogTeamHierarchyLinksExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := teamHierarchyLinksExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func teamHierarchyLinksExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_team-hierarchy-links" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetTeamsApiV2().GetTeamHierarchyLink(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving TeamHierarchyLinks")
		}
	}
	return nil
}
