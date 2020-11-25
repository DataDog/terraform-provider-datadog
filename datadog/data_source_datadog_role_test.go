package datadog

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDatadogRoleDatasource(t *testing.T) {
	accProviders, _, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRoleConfig(),
				Check:  resource.TestCheckResourceAttr("data.datadog_role.foo", "name", "Datadog Standard Role"),
			},
		},
	})
}

func TestAccDatadogRoleDatasourceExactMatch(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	rolename := strings.ToLower(uniqueEntityName(clock, t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDatadogRoleDestroy(accProvider),
		Providers:    accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceRoleCreateConfig(rolename),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_role.main", "name", rolename+" main"),
					resource.TestCheckResourceAttr("datadog_role.cloned", "name", rolename+" main cloned"),
				),
			},
			{
				Config:             testAccDatasourceRoleExactMatchConfig(rolename),
				Check:              resource.TestCheckResourceAttr("data.datadog_role.exact_match", "name", rolename+" main"),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDatadogRoleDatasourceError(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	rolename := strings.ToLower(uniqueEntityName(clock, t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDatadogRoleDestroy(accProvider),
		Providers:    accProviders,
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
