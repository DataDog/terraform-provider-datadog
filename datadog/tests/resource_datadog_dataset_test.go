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
	datasetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDatasetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDataset(datasetName, product),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDatasetExists(providers.frameworkProvider),
					resource.TestCheckTypeSetElemAttr("datadog_dataset.foo", "created_at", "2025-01-01T00:00:00-08:00"),
				),
			},
		},
	})
}

func TestAccDatadogDataset_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	datasetName := uniqueEntityName(ctx, t)

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
						"datadog_dataset.foo", "principals.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_dataset.foo", "principals.*", "team:87bd1a11-b889-4208-8784-17901bc9c54b"),
					resource.TestCheckResourceAttr(
						"datadog_dataset.foo", "product_filters.filters.#", "3"),
					resource.TestCheckResourceAttr(
						"datadog_dataset.foo", "product_filters.filters.*", "@region:us-east-1"),
				),
			},
		},
	})
}

func TestAccDatadogDataset_InvalidInput(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	datasetName := uniqueEntityName(ctx, t)
	invalidProduct := "ci-visibility"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogDataset(datasetName, invalidProduct),
				ExpectError: regexp.MustCompile(invalidProduct),
			},
			{
				Config:      testAccCheckDatadogDatasetInvalidPrincipal(datasetName, product),
				ExpectError: regexp.MustCompile("Invalid principal format"),
			},
		},
	})
}

func TestAccDatadogDataset_EmptyProductFilters(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	datasetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDatasetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDatasetEmptyProductFilters(datasetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDatasetExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_dataset.foo", "name", datasetName),
					resource.TestCheckResourceAttr(
						"datadog_dataset.foo", "product_filters.#", "0"),
				),
			},
		},
	})
}

func TestAccDatadogDataset_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	datasetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDatasetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDataset(datasetName, product),
			},
			{
				ResourceName:      "datadog_dataset.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogDataset(datasetName string, product string) string {
	return fmt.Sprintf(`
		// define product as terraform
		resource "datadog_dataset" "foo" {
    		name = "%s"
    		principals = ["role:44162ede-4f1c-4eb1-91a2-34c674ebd9b9"]
    
			product_filters {
				product = "%s"
				filters = ["@service:web", "@env:prod"]
			}
		}`, datasetName, product)
}

func testAccCheckDatadogDatasetUpdate(datasetName string, product string) string {
	return fmt.Sprintf(`
		resource "datadog_dataset" "foo" {
			name = "%s-updated"
			principals = ["role:44162ede-4f1c-4eb1-91a2-34c674ebd9b9", "team:87bd1a11-b889-4208-8784-17901bc9c54b"]
			
			product_filters {
				product = "%s"
				filters = ["@service:web", "@env:prod", "@region:us-east-1"]
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
				filters = ["@service:web"]
			}
		}`, datasetName, product)
}

func testAccCheckDatadogDatasetEmptyProductFilters(datasetName string) string {
	return fmt.Sprintf(`
	resource "datadog_dataset" "foo" {
		name = "%s"
		principals = ["role:44162ede-4f1c-4eb1-91a2-34c674ebd9b9"]
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
				if r.Type != "datadog_dataset" {
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
