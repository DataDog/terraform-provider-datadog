package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogRoleDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRoleConfig(),
				Check:  resource.TestCheckResourceAttr("data.datadog_role.foo", "name", "Datadog Standard Role"),
			},
		},
	})
}

func TestAccDatadogRoleDatasourceExactMatch(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	rolename := strings.ToLower(uniqueEntityName(ctx, t))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckDatadogRoleDestroy(accProvider),
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRoleCreateConfig(rolename),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_role.main", "name", rolename+" main"),
					resource.TestCheckResourceAttr("datadog_role.cloned", "name", rolename+" main cloned"),
				),
			},
			{
				Config: testAccDatasourceRoleExactMatchConfig(rolename),
				Check:  resource.TestCheckResourceAttr("data.datadog_role.exact_match", "name", rolename+" main"),
			},
		},
	})
}

func TestAccDatadogRoleDatasourceError(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	rolename := strings.ToLower(uniqueEntityName(ctx, t))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckDatadogRoleDestroy(accProvider),
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRoleCreateConfig(rolename),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_role.main", "name", rolename+" main"),
					resource.TestCheckResourceAttr("datadog_role.cloned", "name", rolename+" main cloned"),
				),
			},
			{
				Config:      testAccDatasourceRoleErrorConfig(rolename),
				ExpectError: regexp.MustCompile("no exact match for name .* were found"),
			},
		},
	})
}

func testAccDatasourceRoleConfig() string {
	return `
data "datadog_role" "foo" {
  filter = "Datadog Standard Role"
}`
}

func testAccDatasourceRoleCreateConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_role" "main" {
  name      = "%s main"
}

resource "datadog_role" "cloned" {
  name      = "%s main cloned"
}`, uniq, uniq)

}

func testAccDatasourceRoleErrorConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_role" "error" {
  filter = "%s"
  depends_on = ["datadog_role.main", "datadog_role.cloned"]
}`, testAccDatasourceRoleCreateConfig(uniq), uniq)
}

func testAccDatasourceRoleExactMatchConfig(uniq string) string {
	return fmt.Sprintf(`
%s

data "datadog_role" "exact_match" {
  filter = "%s main"
  depends_on = ["datadog_role.main", "datadog_role.cloned"]
}`, testAccDatasourceRoleCreateConfig(uniq), uniq)
}
