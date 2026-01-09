package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogTeamsDatasourceFilterKeyword(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	namePrefix := "team-" + uniq
	handlePrefix := "team-" + uniq
	filter := namePrefix + "0"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTeamsFilterConfig(filter, namePrefix, handlePrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_teams.teams", "teams.#", "1"),
					resource.TestCheckResourceAttr("data.datadog_teams.teams", "teams.0.name", namePrefix+"01"),
					resource.TestCheckResourceAttr("data.datadog_teams.teams", "teams.0.handle", handlePrefix+"01"),
					resource.TestCheckResourceAttr("data.datadog_teams.teams", "teams.0.user_count", "0"),
				),
			},
		},
	})
}

func testAccDatasourceTeamsFilterConfig(filter, name string, handle string) string {
	return fmt.Sprintf(`
data "datadog_teams" "teams" {
	filter_keyword = "%[1]s"
	depends_on = [
		datadog_team.matching,
		datadog_team.not_matching,
	]
}

resource "datadog_team" "matching" {
  description = "Team description"
  name        = "%[2]s01"
  handle      = "%[3]s01"
}

resource "datadog_team" "not_matching" {
  description = "Team description"
  name        = "%[2]s1"
  handle      = "%[3]s1"
}
`, filter, name, handle)
}
