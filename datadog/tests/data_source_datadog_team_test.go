package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogTeamDatasourceBasic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTeamConfig(uniq),
				Check:  resource.TestCheckResourceAttr("data.datadog_team.foo", "handle", uniq),
			},
		},
	})
}

func testAccDatasourceTeamConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_team" "foo" {
	team_id    = datadog_team.foo.id
}

resource "datadog_team" "foo" {
	description = "Team description"
	handle      = "%s"
	name        = "%s"
}
`, uniq, uniq)
}
