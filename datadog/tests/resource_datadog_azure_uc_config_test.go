package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccAzureUcConfigBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAzureUcConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAzureUcConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAzureUcConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "account_id", "12345678-1234-1234-1234-123456789012"),
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "client_id", "87654321-4321-4321-4321-210987654321"),
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "scope", "/subscriptions/12345678-1234-1234-1234-123456789012"),
					// Check actual_bill_config attributes
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "actual_bill_config.export_name", "test-actual-export"),
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "actual_bill_config.export_path", "/actual-cost"),
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "actual_bill_config.storage_account", "teststorageaccount"),
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "actual_bill_config.storage_container", "test-container"),
					// Check amortized_bill_config attributes
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "amortized_bill_config.export_name", "test-amortized-export"),
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "amortized_bill_config.export_path", "/amortized-cost"),
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "amortized_bill_config.storage_account", "teststorageaccount"),
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "amortized_bill_config.storage_container", "test-container"),
				),
			},
			{
				Config: testAccCheckDatadogAzureUcConfigDataSource(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAzureUcConfigExists(providers.frameworkProvider),
					// Check resource attributes
					resource.TestCheckResourceAttr(
						"datadog_azure_uc_config.foo", "account_id", "12345678-1234-1234-1234-123456789012"),
					// Check data source attributes
					resource.TestCheckResourceAttr(
						"data.datadog_azure_uc_config.bar", "account_id", "12345678-1234-1234-1234-123456789012"),
					resource.TestCheckResourceAttr(
						"data.datadog_azure_uc_config.bar", "client_id", "87654321-4321-4321-4321-210987654321"),
					resource.TestCheckResourceAttr(
						"data.datadog_azure_uc_config.bar", "scope", "/subscriptions/12345678-1234-1234-1234-123456789012"),
					// Check data source nested blocks
					resource.TestCheckResourceAttr(
						"data.datadog_azure_uc_config.bar", "actual_bill_config.export_name", "test-actual-export"),
					resource.TestCheckResourceAttr(
						"data.datadog_azure_uc_config.bar", "actual_bill_config.storage_account", "teststorageaccount"),
					resource.TestCheckResourceAttr(
						"data.datadog_azure_uc_config.bar", "amortized_bill_config.export_name", "test-amortized-export"),
					resource.TestCheckResourceAttr(
						"data.datadog_azure_uc_config.bar", "amortized_bill_config.storage_account", "teststorageaccount"),
					// Verify that data source has the computed fields populated
					resource.TestCheckResourceAttrSet(
						"data.datadog_azure_uc_config.bar", "status"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_azure_uc_config.bar", "created_at"),
				),
			},
		},
	})
}

func TestAccAzureUcConfigImport(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAzureUcConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAzureUcConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAzureUcConfigExists(providers.frameworkProvider),
				),
			},
			{
				ResourceName:      "datadog_azure_uc_config.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogAzureUcConfig(uniq string) string {
	_ = uniq // unused parameter
	return `resource "datadog_azure_uc_config" "foo" {
    account_id = "12345678-1234-1234-1234-123456789012"
    client_id = "87654321-4321-4321-4321-210987654321"
    scope = "/subscriptions/12345678-1234-1234-1234-123456789012"
    
    actual_bill_config {
        export_name = "test-actual-export"
        export_path = "/actual-cost"
        storage_account = "teststorageaccount"
        storage_container = "test-container"
    }
    
    amortized_bill_config {
        export_name = "test-amortized-export"
        export_path = "/amortized-cost"
        storage_account = "teststorageaccount"
        storage_container = "test-container"
    }
}`
}

func testAccCheckDatadogAzureUcConfigDataSource(uniq string) string {
	_ = uniq // unused parameter
	return `
resource "datadog_azure_uc_config" "foo" {
    account_id = "12345678-1234-1234-1234-123456789012"
    client_id = "87654321-4321-4321-4321-210987654321"
    scope = "/subscriptions/12345678-1234-1234-1234-123456789012"
    
    actual_bill_config {
        export_name = "test-actual-export"
        export_path = "/actual-cost"
        storage_account = "teststorageaccount"
        storage_container = "test-container"
    }
    
    amortized_bill_config {
        export_name = "test-amortized-export"
        export_path = "/amortized-cost"
        storage_account = "teststorageaccount"
        storage_container = "test-container"
    }
}

data "datadog_azure_uc_config" "bar" {
    cloud_account_id = datadog_azure_uc_config.foo.id
}`
}

func testAccCheckDatadogAzureUcConfigDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := AzureUcConfigDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func AzureUcConfigDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_azure_uc_config" {
				continue
			}
			id := r.Primary.ID
			cloudAccountId, _ := strconv.ParseInt(id, 10, 64)

			resp, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetCostAzureUCConfig(auth, cloudAccountId)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AzureUcConfig %s", err)}
			}
			// Check if resource is archived (deleted)
			if data, ok := resp.GetDataOk(); ok {
				if attributes, ok := data.GetAttributesOk(); ok {
					if configs, ok := attributes.GetConfigsOk(); ok && len(*configs) > 0 {
						if (*configs)[0].GetStatus() == "archived" {
							return nil // Resource is properly deleted (archived)
						}
					}
				}
			}
			return &utils.RetryableError{Prob: fmt.Sprintf("AzureUcConfig still exists with status other than archived")}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogAzureUcConfigExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := azureUcConfigExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func azureUcConfigExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_azure_uc_config" {
			continue
		}
		id := r.Primary.ID
		cloudAccountId, _ := strconv.ParseInt(id, 10, 64)

		_, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetCostAzureUCConfig(auth, cloudAccountId)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving AzureUcConfig")
		}
	}
	return nil
}
