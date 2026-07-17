package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogRolePermissionsDatasourceBasic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRolePermissionsConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_role_permissions.foo", "permissions.0.permission_id"),
					resource.TestCheckResourceAttrSet("data.datadog_role_permissions.foo", "permissions.0.name"),
				),
			},
		},
	})
}

func testAccDatasourceRolePermissionsConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_permissions" "foo" {}

resource "datadog_role" "uniq_role" {
	name = "%[1]s"
	permission {
		id = data.datadog_permissions.foo.permissions.dashboards_read
	}
}

data "datadog_role_permissions" "foo" {
	role_id = datadog_role.uniq_role.id
}
`, uniq)
}
