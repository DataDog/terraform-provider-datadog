package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogApmServicesDatasource_basic(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	check := func(attr, value string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr("data.datadog_apm_services.test", attr, value)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "datadog_apm_services" "test" {
					filter_env = "*"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					check("id", "*"),
					check("services.#", "3"),
					check("services.0.name", "checkout-service"),
					check("services.0.is_traced", "true"),
					check("services.0.is_usm", "false"),
					check("services.1.name", "payments-service"),
					check("services.1.is_traced", "true"),
					check("services.1.is_usm", "true"),
					check("services.2.name", "web-store"),
					check("services.2.is_traced", "false"),
					check("services.2.is_usm", "false"),
				),
			},
		},
	})
}
