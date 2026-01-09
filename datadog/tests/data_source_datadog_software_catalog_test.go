package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogSoftwareCatalogEntity_Datasource(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Create a service
				Config: testAccDatadogSoftwareCatalogEntitiyResourceConfid(uniq),
				Check:  checkCatalogEntityForDatasourceExists(providers.frameworkProvider, uniq),
			},
			{
				Config: testAccDatadogSoftwareCatalogEntitiesConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_software_catalog.foo", "entities.#", "1"),
					resource.TestCheckResourceAttr("data.datadog_software_catalog.foo", "entities.0.name", uniq),
					resource.TestCheckResourceAttr("data.datadog_software_catalog.foo", "entities.0.kind", "service"),
					resource.TestCheckResourceAttr("data.datadog_software_catalog.foo", "entities.0.owner", fmt.Sprintf("owner.of.%s", uniq)),
				),
			},
		},
	})
}

func checkCatalogEntityForDatasourceExists(accProvider *fwprovider.FrameworkProvider, uniq string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth
		err := utils.Retry(5000*time.Millisecond, 4, func() error {
			optionalParams := datadogV2.NewListCatalogEntityOptionalParameters().WithFilterName(uniq)
			list, _, err := apiInstances.GetSoftwareCatalogApiV2().ListCatalogEntity(auth, *optionalParams)
			if err != nil {
				return err
			}
			if int(*list.Meta.Count) == 0 {
				return fmt.Errorf("no services received from the API after service creation")
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccDatadogSoftwareCatalogEntitiesConfig(uniq string) string {
	return fmt.Sprintf(
		`data "datadog_software_catalog" "foo" {
		filter_name = "%s"
	}`, uniq)
}

func testAccDatadogSoftwareCatalogEntitiyResourceConfid(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_software_catalog" "bar_service" {
		  entity =<<EOF
apiVersion: v3
kind: service
metadata:
  name: %s
  displayName: %s
  tags:
    - tag:value
  owner: owner.of.%s
EOF
			}`, uniq, uniq, uniq)
}
