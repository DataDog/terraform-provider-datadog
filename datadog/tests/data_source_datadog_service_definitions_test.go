package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogServiceDefinitionsDatasource(t *testing.T) {
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceDefinitionsConfig(uniq),
				Check:  resource.TestCheckResourceAttr("data.datadog_service_definitions.foo", "service", uniq),
			},
		},
	})
}

func testAccDatasourceServiceDefinitionsConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_service_definitions" "foo" {
  retrieve_all    = true
  depends_on = [
	  datadog_service_definition_yaml.service_definition_v2_2
  ]
}

resource "datadog_service_definition_yaml" "service_definition_v2_2" {
  service_definition = <<EOF
  schema-version: v2.2
  dd-service: %[1]s
  description: my datadog service 
  tier: high
  lifecycle: production
  application: socialplate
  EOF
`, uniq)
}
