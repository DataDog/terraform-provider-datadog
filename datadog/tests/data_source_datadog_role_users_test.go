package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogRoleUsersDatasourceBasic(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRoleUsersConfig(uniq),
				Check:  resource.TestCheckResourceAttrSet("data.datadog_role_users.foo", "role_users.0.role_id"),
			},
		},
	})
}

func TestAccDatadogRoleUsersDatasourceExactMatch(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRoleUsersExactMatchConfig(uniq, "false"),
				Check:  resource.TestCheckResourceAttr("data.datadog_role_users.ru", "role_users.#", "2"),
			},
			{
				Config: testAccDatasourceRoleUsersExactMatchConfig(uniq, "true"),
				Check:  resource.TestCheckResourceAttr("data.datadog_role_users.ru", "role_users.#", "1"),
			},
		},
	})
}

func testAccDatasourceRoleUsersConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
	email = "%[1]s@example.com"
}

resource "datadog_role" "uniq_role" {
	name = "%[1]s"
}

resource "datadog_user_role" "foo" {
	role_id = datadog_role.uniq_role.id
	user_id = datadog_user.foo.id
}

data "datadog_role_users" "foo" {
	role_id    = datadog_role.uniq_role.id
	depends_on = [ datadog_user_role.foo ]
}
`, uniq)
}

func testAccDatasourceRoleUsersExactMatchConfig(uniq, exactMatch string) string {
	return fmt.Sprintf(`
resource "datadog_role" "uniq_role" {
	name = "%[1]s"
}

resource "datadog_user" "foo" {
	email = "%[1]s@example.com"
	name  = "Foo BarBar"
}

resource "datadog_user" "bar" {
	email = "%[1]s1@example.com"
	name  = "Foo Bar"
}

resource "datadog_user_role" "foo" {
	role_id = datadog_role.uniq_role.id
	user_id = datadog_user.foo.id
}

resource "datadog_user_role" "bar" {
	role_id = datadog_role.uniq_role.id
	user_id = datadog_user.bar.id
}

data "datadog_role_users" "ru" {
	role_id        = datadog_role.uniq_role.id
	exact_match    = %[2]s
	filter         = "Foo Bar"
	depends_on     = [ datadog_user_role.foo, datadog_user_role.bar ]
}
`, uniq, exactMatch)
}
