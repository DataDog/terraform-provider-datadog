package datadog

import (
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

func testAccDatasourceRoleConfig() string {
	return `
data "datadog_role" "foo" {
  filter = "Datadog Standard Role"
}`
}
