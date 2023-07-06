package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogIntegrationAWSNamespaceRulesDatasource(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				data "datadog_integration_aws_namespace_rules" "test" {}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_integration_aws_namespace_rules.test", "namespace_rules.#"),
				),
			},
		},
	})
}
