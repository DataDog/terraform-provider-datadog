package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDatadogPermissionsDatasource(t *testing.T) {
	accProviders, _, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "datadog_permissions" "foo" {}`,
				// Check at least one permission exists
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_permissions.foo", "id", "datadog-permissions"),
					resource.TestCheckResourceAttrSet("data.datadog_permissions.foo", "permissions.admin"),
					resource.TestCheckNoResourceAttr("data.datadog_permissions.foo", "permissions.dashboards_read"),
				),
			},
		},
	})
}
