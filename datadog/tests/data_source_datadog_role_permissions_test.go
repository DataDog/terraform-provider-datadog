package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogRolePermissionsDatasourceBasic(t *testing.T) {
	// Not parallel: the permissions cache reset must not race with other role tests.
	t.Cleanup(datadog.ResetPermissionsCache())
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

	dynamic "permission" {
		for_each = data.datadog_permissions.foo.permissions
		content {
		    id = permission.value
		}
	}
}

data "datadog_role_permissions" "foo" {
	role_id = datadog_role.uniq_role.id
}
`, uniq)
}
