package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The data source lists every RUM application in the org (no server-side
// filter), so we seed one application and assert it appears somewhere in the
// results via a set-style lookup rather than a fixed index.
func TestAccDatadogRumApplicationsDatasource(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	appName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRUMApplicationDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDatadogRumApplicationsConfig(appName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_rum_applications.foo", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_rum_applications.foo", "rum_applications.#"),
					resource.TestCheckTypeSetElemNestedAttrs("data.datadog_rum_applications.foo", "rum_applications.*", map[string]string{
						"name": appName,
						"type": "browser",
					}),
				),
			},
		},
	})
}

func testAccDatasourceDatadogRumApplicationsConfig(appName string) string {
	return fmt.Sprintf(`
resource "datadog_rum_application" "foo" {
	name = "%s"
	type = "browser"
}

data "datadog_rum_applications" "foo" {
	depends_on = [datadog_rum_application.foo]
}
`, appName)
}
