package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Smoke test: the Usage Cost API rejects configs whose Azure credentials cannot be
// verified upstream (AZURE_API_CREDENTIALS_NOT_AVAILABLE), so this test cannot seed
// its own fixture. It reads whatever the test org holds and only checks that the
// list call succeeds and deserializes into state.
func TestAccDatadogAzureUcConfigsDatasource(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "datadog_azure_uc_configs" "foo" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_azure_uc_configs.foo", "id"),
					resource.TestCheckResourceAttrSet("data.datadog_azure_uc_configs.foo", "azure_uc_configs.#"),
				),
			},
		},
	})
}
