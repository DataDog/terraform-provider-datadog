package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogSensitiveDataScannerGroup_SamplingOrder(t *testing.T) {
	cleanupSensitiveDataScannerGroups(t)
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))
	resourceName := "datadog_sensitive_data_scanner_group.sample_group"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSensitiveDataScannerGroupDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSensitiveDataScannerGroupAllProductsSampling(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerGroupExists(accProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "product_list.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "samplings.#", "4"),
					testAccCheckDatadogSensitiveDataScannerGroupSampling(resourceName, "logs", 25),
					testAccCheckDatadogSensitiveDataScannerGroupSampling(resourceName, "rum", 25),
					testAccCheckDatadogSensitiveDataScannerGroupSampling(resourceName, "events", 25),
					testAccCheckDatadogSensitiveDataScannerGroupSampling(resourceName, "apm", 25),
				),
			},
		},
	})
}

func TestAccDatadogSensitiveDataScannerGroup_Basic(t *testing.T) {
	//if isRecording() || isReplaying() {
	//	t.Skip("This test doesn't support recording or replaying")
	//}
	cleanupSensitiveDataScannerGroups(t)
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))
	resource_name := "datadog_sensitive_data_scanner_group.sample_group"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSensitiveDataScannerGroupDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSensitiveDataScannerGroup(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerGroupExists(accProvider, resource_name),
					resource.TestCheckResourceAttr(resource_name, "name", uniq),
					resource.TestCheckResourceAttr(resource_name, "description", ""),
					resource.TestCheckResourceAttr(resource_name, "product_list.0", "logs"),
					resource.TestCheckResourceAttr(resource_name, "product_list.#", "1"),
					resource.TestCheckResourceAttr(resource_name, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resource_name, "filter.0.query", "*"),
					resource.TestCheckResourceAttr(resource_name, "samplings.0.product", "logs"),
					resource.TestCheckResourceAttr(resource_name, "samplings.0.rate", "100"),
				),
			},
			{
				Config: testAccCheckDatadogSensitiveDataScannerGroupUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerGroupExists(accProvider, resource_name),
					resource.TestCheckResourceAttr(resource_name, "name", uniq),
					resource.TestCheckResourceAttr(resource_name, "description", "changed description"),
					resource.TestCheckResourceAttr(resource_name, "product_list.#", "2"),
					resource.TestCheckResourceAttr(resource_name, "is_enabled", "false"),
					resource.TestCheckResourceAttr(resource_name, "filter.0.query", "hotel:trivago2.0"),
					resource.TestCheckResourceAttr(resource_name, "samplings.0.product", "logs"),
					resource.TestCheckResourceAttr(resource_name, "samplings.0.rate", "100"),
					resource.TestCheckResourceAttr(resource_name, "samplings.1.product", "apm"),
					resource.TestCheckResourceAttr(resource_name, "samplings.1.rate", "10"),
				),
			},
		},
	})
}

func testAccCheckDatadogSensitiveDataScannerGroupAllProductsSampling(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name         = "%s"
	product_list = ["logs", "rum", "events", "apm"]
	is_enabled   = true
	filter {
		query = "service:sds-scenario-runner"
	}
	samplings {
		product = "logs"
		rate    = 25
	}
	samplings {
		product = "rum"
		rate    = 25
	}
	samplings {
		product = "events"
		rate    = 25
	}
	samplings {
		product = "apm"
		rate    = 25
	}
}
`, name)
}

func testAccCheckDatadogSensitiveDataScannerGroupSampling(resourceName, product string, rate float64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		count, err := strconv.Atoi(resourceState.Primary.Attributes["samplings.#"])
		if err != nil {
			return fmt.Errorf("failed to parse samplings count: %w", err)
		}

		for i := 0; i < count; i++ {
			prefix := fmt.Sprintf("samplings.%d.", i)
			if resourceState.Primary.Attributes[prefix+"product"] == product {
				rateStr := resourceState.Primary.Attributes[prefix+"rate"]
				if rateStr == strconv.FormatFloat(rate, 'f', -1, 64) {
					return nil
				}
				return fmt.Errorf("sampling for product %q has rate %q, want %v", product, rateStr, rate)
			}
		}

		return fmt.Errorf("sampling for product %q not found in state", product)
	}
}

func testAccCheckDatadogSensitiveDataScannerGroup(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name           = "%s"
	product_list   = ["logs"]
	is_enabled     = true
	filter {
		query = ""
	}
	samplings {
		product = "logs"
		rate    = 100
	}
}
`, name)
}

func testAccCheckDatadogSensitiveDataScannerGroupUpdate(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name           = "%s"
	description    = "changed description"
	product_list   = ["logs", "apm"]
	is_enabled     = false
	filter {
		query = "hotel:trivago2.0"
	}
	samplings {
		product = "logs"
		rate    = 100
	}
	samplings {
		product = "apm"
		rate    = 10
	}
}
`, name)
}

func testAccCheckDatadogSensitiveDataScannerGroupExists(accProvider func() (*schema.Provider, error), name string) resource.TestCheckFunc {
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

func testAccCheckDatadogSensitiveDataScannerGroupDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, r := range s.RootModule().Resources {
			if r.Type == "datadog_sensitive_data_scanner_group" {
				resp, _, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
				if groupFound := findSensitiveDataScannerGroupHelper(r.Primary.ID, resp); groupFound == nil {
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
	for _, r := range response.GetIncluded() {
		if r.SensitiveDataScannerGroupIncludedItem.GetId() == groupId {
			return r.SensitiveDataScannerGroupIncludedItem
		}
	}

	return nil
}

func TestAccDatadogSensitiveDataScannerGroup_DeleteAlreadyDeleted(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	cleanupSensitiveDataScannerGroups(t)
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))
	var groupId string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSensitiveDataScannerGroupDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSensitiveDataScannerGroup(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerGroupExists(accProvider, "datadog_sensitive_data_scanner_group.sample_group"),
					// Capture the group ID for deletion
					func(s *terraform.State) error {
						groupId = s.RootModule().Resources["datadog_sensitive_data_scanner_group.sample_group"].Primary.ID
						return nil
					},
				),
			},
			{
				// Delete the group via API before Terraform tries to destroy it
				PreConfig: func() {
					provider, _ := accProvider()
					providerConf := provider.Meta().(*datadog.ProviderConfiguration)
					apiInstances := providerConf.DatadogApiInstances
					auth := providerConf.Auth

					body := datadogV2.NewSensitiveDataScannerGroupDeleteRequestWithDefaults()
					metaVar := datadogV2.NewSensitiveDataScannerMetaVersionOnlyWithDefaults()
					body.SetMeta(*metaVar)
					_, _, err := apiInstances.GetSensitiveDataScannerApiV2().DeleteScanningGroup(auth, groupId, *body)
					if err != nil {
						t.Logf("Warning: failed to delete group via API: %v", err)
					}
				},
				// Empty config to trigger destroy - should succeed even though resource is already deleted
				Config: `# Empty config`,
			},
		},
	})
}
