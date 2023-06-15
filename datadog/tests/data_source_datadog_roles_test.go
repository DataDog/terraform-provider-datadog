package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogRolesDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRolesConfig("Datadog Admin Role"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_roles.foo", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_roles.foo", "roles.0.id"),
					resource.TestCheckResourceAttr("data.datadog_roles.foo", "roles.0.name", "Datadog Admin Role"),
				),
			},
		},
	})
}

func TestAccDatadogRolesDatasourceMultipleMatch(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	rolename := strings.ToLower(uniqueEntityName(ctx, t))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckDatadogRoleDestroy(accProvider),
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRolesMultipleMatchConfig(rolename),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_roles.foo", "roles.#", "2"),
					resource.TestCheckResourceAttrSet("data.datadog_roles.foo", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_roles.foo", "roles.0.id"),
					resource.TestCheckResourceAttr("data.datadog_roles.foo", "roles.0.name", rolename+" main"),
					resource.TestCheckResourceAttrSet("data.datadog_roles.foo", "roles.1.id"),
					resource.TestCheckResourceAttr("data.datadog_roles.foo", "roles.1.name", rolename+" main cloned"),
				),
			},
		},
	})
}

func testAccDatasourceRolesConfig(filter string) string {
	return fmt.Sprintf(`
data "datadog_roles" "foo" {
  filter = "%s"
}`, filter)
}

func testAccDatasourceRolesMultipleMatchConfig(filter string) string {
	return fmt.Sprintf(`
data "datadog_roles" "foo" {
  filter = "%[1]s"
  depends_on = [datadog_role.main, datadog_role.cloned]
}

resource "datadog_role" "main" {
  name      = "%[1]s main"
}

resource "datadog_role" "cloned" {
  name      = "%[1]s main cloned"
}`, filter)
}
