package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamPermissionSettingBasic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTeamPermissionSetting(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_team_permission_setting.foo", "action", "edit"),
					resource.TestCheckResourceAttr(
						"datadog_team_permission_setting.foo", "value", "teams_manage"),
				),
			},
			{
				Config: testAccCheckDatadogTeamPermissionSettingUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_team_permission_setting.foo", "action", "edit"),
					resource.TestCheckResourceAttr(
						"datadog_team_permission_setting.foo", "value", "members"),
				),
			},
		},
	})
}

func testAccCheckDatadogTeamPermissionSetting(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_team" "foo" {
	description = "Example team"
	handle      = "%s"
	name        = "%s"
}

resource "datadog_team_permission_setting" "foo" {
	team_id        = datadog_team.foo.id
	action         = "edit"
	value          = "teams_manage"
  }  
`, uniq, uniq)
}

func testAccCheckDatadogTeamPermissionSettingUpdated(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_team" "foo" {
	description = "Example team"
	handle      = "%s"
	name        = "%s"
}

resource "datadog_team_permission_setting" "foo" {
	team_id        = datadog_team.foo.id
	action         = "edit"
	value          = "members"
  }  
`, uniq, uniq)
}
