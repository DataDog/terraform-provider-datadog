package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogCloudInventorySyncConfig_AWS(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCloudInventorySyncConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudInventorySyncConfigAWS(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCloudInventorySyncConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.aws_test", "cloud_provider", "aws"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.aws_test", "aws_account_id", "123456789012"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.aws_test", "destination_bucket_name", fmt.Sprintf("test-bucket-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.aws_test", "destination_bucket_region", "us-east-1"),
				),
			},
			{
				Config: testAccCheckDatadogCloudInventorySyncConfigAWSUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCloudInventorySyncConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.aws_test", "cloud_provider", "aws"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.aws_test", "destination_bucket_region", "us-west-2"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.aws_test", "destination_prefix", "inventory/"),
				),
			},
		},
	})
}

func TestAccDatadogCloudInventorySyncConfig_Azure(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCloudInventorySyncConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudInventorySyncConfigAzure(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCloudInventorySyncConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.azure_test", "cloud_provider", "azure"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.azure_test", "azure_client_id", "11111111-1111-1111-1111-111111111111"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.azure_test", "azure_tenant_id", "22222222-2222-2222-2222-222222222222"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.azure_test", "azure_subscription_id", "33333333-3333-3333-3333-333333333333"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.azure_test", "azure_resource_group", "test-rg"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.azure_test", "azure_storage_account", fmt.Sprintf("teststorage%s", uniq)),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.azure_test", "azure_container", "inventory"),
				),
			},
		},
	})
}

func TestAccDatadogCloudInventorySyncConfig_GCP(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCloudInventorySyncConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudInventorySyncConfigGCP(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCloudInventorySyncConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.gcp_test", "cloud_provider", "gcp"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.gcp_test", "gcp_project_id", fmt.Sprintf("test-project-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.gcp_test", "gcp_destination_bucket_name", fmt.Sprintf("dest-bucket-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.gcp_test", "gcp_source_bucket_name", fmt.Sprintf("source-bucket-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.gcp_test", "gcp_service_account_email", fmt.Sprintf("sa@test-project-%s.iam.gserviceaccount.com", uniq)),
				),
			},
		},
	})
}

func TestAccDatadogCloudInventorySyncConfig_InvalidProvider(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogCloudInventorySyncConfigInvalidProvider(),
				ExpectError: regexp.MustCompile(`Attribute cloud_provider value must be one of`),
			},
		},
	})
}

func TestAccDatadogCloudInventorySyncConfig_MissingRequiredAWSFields(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogCloudInventorySyncConfigMissingAWSFields(),
				ExpectError: regexp.MustCompile(`Missing required field`),
			},
		},
	})
}

func testAccCheckDatadogCloudInventorySyncConfigExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_cloud_inventory_sync_config" {
				continue
			}

			_, _, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", "/api/unstable/cloudinventoryservice/syncstatus", nil)
			if err != nil {
				return fmt.Errorf("error retrieving cloud inventory sync config: %s", err)
			}
		}
		return nil
	}
}

func testAccCheckDatadogCloudInventorySyncConfigDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		// Cloud Inventory Sync Config cannot be deleted from Datadog
		// It is only removed from Terraform state, so we don't check for destruction
		return nil
	}
}

// AWS configuration
func testAccCheckDatadogCloudInventorySyncConfigAWS(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_inventory_sync_config" "aws_test" {
  cloud_provider          = "aws"
  aws_account_id          = "123456789012"
  destination_bucket_name = "test-bucket-%s"
  destination_bucket_region = "us-east-1"
}`, uniq)
}

func testAccCheckDatadogCloudInventorySyncConfigAWSUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_inventory_sync_config" "aws_test" {
  cloud_provider          = "aws"
  aws_account_id          = "123456789012"
  destination_bucket_name = "test-bucket-%s"
  destination_bucket_region = "us-west-2"
  destination_prefix      = "inventory/"
}`, uniq)
}

// Azure configuration
func testAccCheckDatadogCloudInventorySyncConfigAzure(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_inventory_sync_config" "azure_test" {
  cloud_provider        = "azure"
  azure_client_id       = "11111111-1111-1111-1111-111111111111"
  azure_tenant_id       = "22222222-2222-2222-2222-222222222222"
  azure_subscription_id = "33333333-3333-3333-3333-333333333333"
  azure_resource_group  = "test-rg"
  azure_storage_account = "teststorage%s"
  azure_container       = "inventory"
}`, uniq)
}

// GCP configuration
func testAccCheckDatadogCloudInventorySyncConfigGCP(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_inventory_sync_config" "gcp_test" {
  cloud_provider             = "gcp"
  gcp_project_id             = "test-project-%s"
  gcp_destination_bucket_name = "dest-bucket-%s"
  gcp_source_bucket_name     = "source-bucket-%s"
  gcp_service_account_email  = "sa@test-project-%s.iam.gserviceaccount.com"
}`, uniq, uniq, uniq, uniq)
}

// Invalid configurations
func testAccCheckDatadogCloudInventorySyncConfigInvalidProvider() string {
	return `
resource "datadog_cloud_inventory_sync_config" "invalid" {
  cloud_provider = "invalid"
}`
}

func testAccCheckDatadogCloudInventorySyncConfigMissingAWSFields() string {
	return `
resource "datadog_cloud_inventory_sync_config" "missing_fields" {
  cloud_provider = "aws"
  aws_account_id = "123456789012"
  # Missing destination_bucket_name and destination_bucket_region
}`
}
