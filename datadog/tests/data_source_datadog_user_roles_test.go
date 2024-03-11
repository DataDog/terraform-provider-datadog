package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogUserRolesDatasourceBasic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserRolesConfig(uniq),
				Check:  resource.TestCheckResourceAttr("data.datadog_user_roles.foo", "user_roles.0.role", "admin"),
			},
		},
	})
}

func TestAccDatadogUserRolesDatasourceExactMatch(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserRolesExactMatchConfig(uniq, "false"),
				Check:  resource.TestCheckResourceAttr("data.datadog_user_roles.foo", "user_roles.#", "2"),
			},
			{
				Config: testAccDatasourceUserRolesExactMatchConfig(uniq, "true"),
				Check:  resource.TestCheckResourceAttr("data.datadog_user_roles.foo", "user_roles.#", "1"),
			},
		},
	})
}

func testAccDatasourceUserRolesConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_user_roles" "foo" {
	role_id    = datadog_role.foo.id
	depends_on = [ datadog_user_role.foo ]
}

resource "datadog_user" "foo" {
	email = "%s@example.com"
}

data "datadog_role" "std_role" {
	filter = "Datadog Standard Role"
}

resource "datadog_user_role" "foo" {
	role_id = datadog_role.std_role.id
	user_id = datadog_user.foo.id
}
`, uniq)
}

func testAccDatasourceUserRolesExactMatchConfig(uniq, exactMatch string) string {
	return fmt.Sprintf(`
data "datadog_user_roles" "foo" {
	role_id        = datadog_role.std_role.id
	exact_match    = %[2]s
	filter_keyword = "Foo Bar"
	depends_on     = [ datadog_user_role.foo, datadog_user_role.bar ]
}

resource "datadog_user" "foo" {
	email = "%[1]s@example.com"
	name  = "Foo BarBar"
}

resource "datadog_user" "bar" {
	email = "%[1]s1@example.com"
	name  = "Foo Bar"
}

data "datadog_role" "std_role" {
	filter = "Datadog Standard Role"
}

resource "datadog_user_role" "foo" {
	role_id = datadog_role.std_role.id
	user_id = datadog_user.foo.id
}

resource "datadog_user_role" "bar" {
	role_id = datadog_role.std_role.id
	user_id = datadog_user.bar.id
}
`, uniq, exactMatch)
}
