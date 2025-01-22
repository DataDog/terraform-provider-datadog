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
	name := "team-" + uniq
	handle := "team-" + uniq

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTeamsFilterConfig(uniq, name, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_teams.teams", "teams.#", "1"),
					resource.TestCheckResourceAttr("data.datadog_teams.teams", "teams.0.name", name),
					resource.TestCheckResourceAttr("data.datadog_teams.teams", "teams.0.handle", handle),
					resource.TestCheckResourceAttr("data.datadog_teams.teams", "teams.0.user_count", "0"),
				),
			},
		},
	})
}

func testAccDatasourceTeamsFilterConfig(uniq, name string, handle string) string {
	return fmt.Sprintf(`
data "datadog_teams" "teams" {
	filter_keyword = "%[1]s"
	depends_on = [
		datadog_team.team_0
	]
}

resource "datadog_team" "team_0" {
  description = "Team description"
  name        = "%[2]s"
  handle      = "%[3]s"
}
`, uniq, name, handle)
}
