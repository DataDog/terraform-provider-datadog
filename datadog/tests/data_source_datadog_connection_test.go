package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestConns(t *testing.T) {
	t.Parallel()

	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionID := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testConnectionDataSourceConfig(connectionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_connection.conn", "name", "test"),
				),
			},
		},
	})
}

func testConnectionDataSourceConfig(uniq string) string {
	return fmt.Sprintf(`
	%s
	data "datadog_connection" "conn" {
		id = "%s"
	}`, testConnectionResourceConfig(), uniq)
}

func testConnectionResourceConfig() string {
	return `
	resource "datadog_connection" "conn" {
		name = "test"

		aws {
			assume_role {
				account_id = "123"
				role = "role"
			}
		}
	}`
}
