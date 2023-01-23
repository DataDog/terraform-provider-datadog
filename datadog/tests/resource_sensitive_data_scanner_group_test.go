package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

func TestAccDatadogSensitiveDataScannerGroup_Basic(t *testing.T) {
	// if isRecording() || isReplaying() {
	// 	t.Skip("This test doesn't support recording or replaying")
	// }
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))
	resource_name := "datadog_sensitive_data_scanner_group.sample_group"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogScanningGroupDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateScanningGroupConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSDSGroupExists(accProvider, resource_name),
					resource.TestCheckResourceAttr(resource_name, "name", uniq),
					resource.TestCheckResourceAttr(resource_name, "description", "optional description"),
					resource.TestCheckResourceAttr(resource_name, "product_list.0", "logs"),
					resource.TestCheckResourceAttr(resource_name, "product_list.#", "1"),
					resource.TestCheckResourceAttr(resource_name, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resource_name, "filter.0.query", "hotel:trivago"),
				),
			},
			{
				Config: testAccCheckDatadogUpdateScanningGroupConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSDSGroupExists(accProvider, resource_name),
					resource.TestCheckResourceAttr(resource_name, "name", uniq),
					resource.TestCheckResourceAttr(resource_name, "description", "changed description"),
					resource.TestCheckResourceAttr(resource_name, "product_list.0", "logs"),
					resource.TestCheckResourceAttr(resource_name, "product_list.1", "apm"),
					resource.TestCheckResourceAttr(resource_name, "product_list.#", "2"),
					resource.TestCheckResourceAttr(resource_name, "is_enabled", "false"),
					resource.TestCheckResourceAttr(resource_name, "filter.0.query", "hotel:trivago2.0"),
				),
			},
		},
	})
}

func testAccCheckDatadogCreateScanningGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
  name           = "%s"
  description    = "optional description"
  product_list   = ["logs"]
  is_enabled     = true
  filter {
	query = "hotel:trivago"
  }
}
`, name)
}

func testAccCheckDatadogUpdateScanningGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name           = "%s"
	description    = "changed description"
	product_list   = ["logs", "apm"]
	is_enabled     = false
	filter {
		query = "hotel:trivago2.0"
	}
}
`, name)
}

func testAccCheckDatadogSDSGroupExists(accProvider func() (*schema.Provider, error), name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		groupId := s.RootModule().Resources[name].Primary.ID
		resp, _, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
		if err != nil {
			return fmt.Errorf("received an error retrieving the list of scanning groups, %s", err)
		}

		if groupFound := findSensitiveDataScannerGroupHelper(groupId, resp); groupFound == nil {
			return fmt.Errorf("received an error retrieving scanning group")
		}

		return nil
	}
}

func testAccCheckDatadogScanningGroupDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_sensitive_data_scanner_group" {
				resp, _, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
				if groupFound := findSensitiveDataScannerGroupHelper(resource.Primary.ID, resp); groupFound == nil {
					if err != nil {
						return fmt.Errorf("received an error retrieving all scanning groups: %s", err)
					}
					return nil
				}
				return fmt.Errorf("scanning group still exists")
			}
		}
		return nil
	}
}

func findSensitiveDataScannerGroupHelper(groupId string, response datadogV2.SensitiveDataScannerGetConfigResponse) *datadogV2.SensitiveDataScannerGroupIncludedItem {
	for _, resource := range response.Included {
		if *resource.SensitiveDataScannerGroupIncludedItem.Id == groupId {
			return resource.SensitiveDataScannerGroupIncludedItem
		}
	}

	return nil
}
