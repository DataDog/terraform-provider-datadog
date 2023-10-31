package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogTeamMembershipsDatasourceBasic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTeamMembershipsConfig(uniq),
				Check:  resource.TestCheckResourceAttr("data.datadog_team_memberships.foo", "team_memberships.0.role", "admin"),
			},
		},
	})
}

func TestAccDatadogTeamMembershipsDatasourceExactMatch(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTeamMembershipsExactMatchConfig(uniq, "false"),
				Check:  resource.TestCheckResourceAttr("data.datadog_team_memberships.foo", "team_memberships.#", "2"),
			},
			{
				Config: testAccDatasourceTeamMembershipsExactMatchConfig(uniq, "true"),
				Check:  resource.TestCheckResourceAttr("data.datadog_team_memberships.foo", "team_memberships.#", "1"),
			},
		},
	})
}

func testAccDatasourceTeamMembershipsConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_team_memberships" "foo" {
	team_id    = datadog_team.foo.id
	depends_on = [ datadog_team_membership.foo ]
}

resource "datadog_user" "foo" {
	email = "%s@example.com"
}

resource "datadog_team" "foo" {
	description = "TeamMemberships description"
	handle      = "%s"
	name        = "%s"
}

resource "datadog_team_membership" "foo" {
	team_id = datadog_team.foo.id
	user_id = datadog_user.foo.id
	role    = "admin"
}
`, uniq, uniq, uniq)
}

func testAccDatasourceTeamMembershipsExactMatchConfig(uniq, exactMatch string) string {
	return fmt.Sprintf(`
data "datadog_team_memberships" "foo" {
	team_id        = datadog_team.foo.id
	exact_match    = %[2]s
	filter_keyword = "Foo Bar"
	depends_on     = [ datadog_team_membership.foo, datadog_team_membership.bar ]
}

resource "datadog_user" "foo" {
	email = "%[1]s@example.com"
	name  = "Foo BarBar"
}

resource "datadog_user" "bar" {
	email = "%[1]s1@example.com"
	name  = "Foo Bar"
}

resource "datadog_team" "foo" {
	description = "TeamMemberships description"
	handle      = "%[1]s"
	name        = "%[1]s"
}

resource "datadog_team_membership" "foo" {
	team_id = datadog_team.foo.id
	user_id = datadog_user.foo.id
	role    = "admin"
}

resource "datadog_team_membership" "bar" {
	team_id = datadog_team.foo.id
	user_id = datadog_user.bar.id
	role    = "admin"
}
`, uniq, exactMatch)
}
