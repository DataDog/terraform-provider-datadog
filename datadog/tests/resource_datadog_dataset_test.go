package test

import (
	"context"
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

func TestAccDatadogDataset_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	datasetName := uniqueDatasetName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDatasetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDataset(datasetName, product),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDatasetExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func TestAccDatadogDataset_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	datasetName := uniqueDatasetName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDatasetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDataset(datasetName, product),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDatasetExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_dataset.foo", "name", datasetName),
				),
			},
			{
				Config: testAccCheckDatadogDatasetUpdate(datasetName, product),
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
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	datasetName := uniqueDatasetName(ctx, t)
	invalidProduct := "ci-visibility"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogDataset(datasetName, invalidProduct),
				ExpectError: regexp.MustCompile("Invalid product"),
			},
			{
				Config:      testAccCheckDatadogDatasetInvalidPrincipal(datasetName, product),
				ExpectError: regexp.MustCompile("PrincipalInvalid"),
			},
			{
				Config:      testAccCheckDatadogDatasetEmptyProductFilters(datasetName),
				ExpectError: regexp.MustCompile("DatasetFiltersEmpty"),
			},
		},
	})
}

func testAccCheckDatadogDataset(datasetName string, product string) string {
	return fmt.Sprintf(`
		resource "datadog_dataset" "foo" {
    		name = "%s"
    		principals = ["role:94172442-be03-11e9-a77a-3b7612558ac1"]
    
			product_filters {
				product = "%s"
				filters = ["@application.id:ce9843b0-7a45-453c-a831-55dd15f85141"]
			}
		}`, datasetName, product)
}

func testAccCheckDatadogDatasetUpdate(datasetName string, product string) string {
	return fmt.Sprintf(`
		resource "datadog_dataset" "foo" {
			name = "%s-updated"
			principals = ["role:94172442-be03-11e9-a77a-3b7612558ac1", "team:4ca6f4c0-88e4-4d42-b7bd-dea73da5c59e"]
			
			product_filters {
				product = "%s"
				filters = ["@application.id:ce9843b0-7a45-453c-a831-55dd15f85141"]
			}

			product_filters {
				product = "apm"
				filters = ["service:test"]
			}
		}`, datasetName, product)
}

func testAccCheckDatadogDatasetInvalidPrincipal(datasetName string, product string) string {
	return fmt.Sprintf(`
		resource "datadog_dataset" "foo" {
			name = "%s"
			principals = ["foo:invalid-principal-format"]
			
			product_filters {
				product = "%s"
				filters = ["@application.id:ce9843b0-7a45-453c-a831-55dd15f85141"]
			}
		}`, datasetName, product)
}

func testAccCheckDatadogDatasetEmptyProductFilters(datasetName string) string {
	return fmt.Sprintf(`
	resource "datadog_dataset" "foo" {
		name = "%s"
		principals = ["role:94172442-be03-11e9-a77a-3b7612558ac1"]
	}`, datasetName)
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

		err := utils.Retry(2, 10, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "resource_datadog_dataset" {
					continue
				}
				id := r.Primary.ID

				_, httpResp, err := apiInstances.GetDatasetsApiV2().GetDataset(auth, id)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
						return nil
					}
					return utils.TranslateClientError(err, httpResp, "error retrieving dataset")
				}
				return fmt.Errorf("dataset still exists")
			}
			return nil
		})
		return err
	}
}
