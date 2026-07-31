package test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	product = "rum"
)

// uniqueDatasetFilters returns the RUM and APM filters to use for a dataset test.
// The API rejects a dataset whose filter is already claimed by another dataset of
// the same product (`DatasetFilterInUse`), so every test must use its own filter
// values, otherwise the dataset tests collide with each other when run in parallel.
// The values are derived from the unique dataset name so they stay stable under the
// frozen clock used for cassette replay.
func uniqueDatasetFilters(ctx context.Context, t *testing.T) (string, string) {
	name := uniqueDatasetName(ctx, t)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(name)))
	applicationID := fmt.Sprintf("%s-%s-%s-%s-%s", hash[0:8], hash[8:12], hash[12:16], hash[16:20], hash[20:32])
	return fmt.Sprintf("@application.id:%s", applicationID), fmt.Sprintf("service:%s", name)
}

func TestAccDatadogDataset_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(t.Context(), t)
	datasetName := uniqueDatasetName(ctx, t)
	rumFilter, _ := uniqueDatasetFilters(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDatasetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDataset(datasetName, product, rumFilter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDatasetExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func TestAccDatadogDataset_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(t.Context(), t)
	datasetName := uniqueDatasetName(ctx, t)
	rumFilter, apmFilter := uniqueDatasetFilters(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDatasetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDataset(datasetName, product, rumFilter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDatasetExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_dataset.foo", "name", datasetName),
				),
			},
			{
				Config: testAccCheckDatadogDatasetUpdate(datasetName, product, rumFilter, apmFilter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDatasetExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_dataset.foo", "name", fmt.Sprintf("%s-updated", datasetName)),
					resource.TestCheckResourceAttr(
						"datadog_dataset.foo", "principals.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_dataset.foo", "product_filters.#", "2"),
				),
			},
		},
	})
}

func TestAccDatadogDataset_InvalidInput(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(t.Context(), t)
	datasetName := uniqueDatasetName(ctx, t)
	rumFilter, _ := uniqueDatasetFilters(ctx, t)
	invalidProduct := "ci-visibility"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDataset(datasetName, invalidProduct, rumFilter),
				// Terraform hard-wraps diagnostic text, so the message may
				// contain a newline between the two words.
				ExpectError: regexp.MustCompile(`Invalid\s+product`),
			},
			{
				Config:      testAccCheckDatadogDatasetInvalidPrincipal(datasetName, product, rumFilter),
				ExpectError: regexp.MustCompile("PrincipalInvalid"),
			},
			{
				Config:      testAccCheckDatadogDatasetEmptyProductFilters(datasetName),
				ExpectError: regexp.MustCompile("DatasetFiltersEmpty"),
			},
		},
	})
}

func testAccCheckDatadogDataset(datasetName string, product string, rumFilter string) string {
	return fmt.Sprintf(`
		resource "datadog_dataset" "foo" {
    		name = "%s"
    		principals = ["role:94172442-be03-11e9-a77a-3b7612558ac1"]

			product_filters {
				product = "%s"
				filters = ["%s"]
			}
		}`, datasetName, product, rumFilter)
}

func testAccCheckDatadogDatasetUpdate(datasetName string, product string, rumFilter string, apmFilter string) string {
	return fmt.Sprintf(`
		resource "datadog_dataset" "foo" {
			name = "%s-updated"
			principals = ["role:94172442-be03-11e9-a77a-3b7612558ac1", "team:4ca6f4c0-88e4-4d42-b7bd-dea73da5c59e"]

			product_filters {
				product = "%s"
				filters = ["%s"]
			}

			product_filters {
				product = "apm"
				filters = ["%s"]
			}
		}`, datasetName, product, rumFilter, apmFilter)
}

func testAccCheckDatadogDatasetInvalidPrincipal(datasetName string, product string, rumFilter string) string {
	return fmt.Sprintf(`
		resource "datadog_dataset" "foo" {
			name = "%s"
			principals = ["foo:invalid-principal-format"]

			product_filters {
				product = "%s"
				filters = ["%s"]
			}
		}`, datasetName, product, rumFilter)
}

func testAccCheckDatadogDatasetEmptyProductFilters(datasetName string) string {
	return fmt.Sprintf(`
	resource "datadog_dataset" "foo" {
		name = "%s"
		principals = ["role:94172442-be03-11e9-a77a-3b7612558ac1"]
	}`, datasetName)
}

func TestAccDatadogDatasetImport(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(t.Context(), t)
	datasetName := uniqueDatasetName(ctx, t)
	rumFilter, _ := uniqueDatasetFilters(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDatasetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDataset(datasetName, product, rumFilter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDatasetExists(providers.frameworkProvider),
				),
			},
			{
				ResourceName:      "datadog_dataset.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogDatasetExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_dataset" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetDatasetsApiV2().GetDataset(auth, id)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving dataset")
			}
		}
		return nil
	}
}

func testAccCheckDatadogDatasetDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_dataset" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetDatasetsApiV2().GetDataset(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
					continue
				}
				return utils.TranslateClientError(err, httpResp, "error retrieving dataset")
			}
			return errors.New("dataset still exists")
		}
		return nil
	}
}
