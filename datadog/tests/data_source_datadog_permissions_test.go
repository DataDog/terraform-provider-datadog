package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDatadogPermissionsDatasource(t *testing.T) {
	t.Parallel()
	ctx := testSpan(context.Background(), t)
	ctx, accProviders := testAccProviders(ctx, t, initRecorder(t))

	resource.Test(t, resource.TestCase{
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
