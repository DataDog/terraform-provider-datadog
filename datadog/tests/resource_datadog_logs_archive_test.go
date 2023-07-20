package test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Test
// create: OK azure
func archiveAzureConfigForCreation(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_azure" "an_azure_integration" {
  tenant_name   = "%s"
  client_id     = "a75fbdd2-ade6-43d0-a810-4d886c53871e"
  client_secret = "testingx./Sw*g/Y33t..R1cH+hScMDt"
}

resource "datadog_logs_archive" "my_azure_archive" {
  depends_on = ["datadog_integration_azure.an_azure_integration"]
  name  = "my first azure archive"
  query = "service:toto"
  azure_archive {
    container 		= "my-container"
    tenant_id 		= "%s"
    client_id       = "a75fbdd2-ade6-43d0-a810-4d886c53871e"
    storage_account = "storageaccount"
    path            = "/path/blou"
  }
}`, uniq, uniq)
}

func TestAccDatadogLogsArchiveAzure_basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	unique_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniqueEntityName(ctx, t))))
	tenantName := fmt.Sprintf("%s-%s-%s-%s-%s", unique_hash[:8], unique_hash[8:12], unique_hash[12:16], unique_hash[16:20], unique_hash[20:32])
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckArchiveAndIntegrationAzureDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveAzureConfigForCreation(tenantName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "name", "my first azure archive"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "query", "service:toto"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure_archive.0.container", "my-container"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure_archive.0.client_id", "a75fbdd2-ade6-43d0-a810-4d886c53871e"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure_archive.0.tenant_id", tenantName),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure_archive.0.storage_account", "storageaccount"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure_archive.0.path", "/path/blou"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "include_tags", "false"),
					resource.TestCheckNoResourceAttr(
						"datadog_logs_archive.my_azure_archive", "rehydration_max_scan_size_in_gb"),
				),
			},
		},
	})
}

// create: Ok gcs
func archiveGCSConfigForCreation(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id     = "%s"
  private_key_id = "1234567890123456789012345678901234567890"
  private_key    = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  client_email   = "%s@awesome-project-id.iam.gserviceaccount.com"
  client_id      = "123456789012345678901"
  host_filters   = "foo:bar,buzz:lightyear"
}

resource "datadog_logs_archive" "my_gcs_archive" {
  depends_on = ["datadog_integration_gcp.awesome_gcp_project_integration"]
  name       = "my first gcs archive"
  query      = "service:tata"
  gcs_archive {
    bucket 		 = "dd-logs-test-datadog-api-client-go"
	path 	     = "/path/blah"
	client_email = "%s@awesome-project-id.iam.gserviceaccount.com"
	project_id   = "%s"
  }
}`, uniq, uniq, uniq, uniq)
}

func TestAccDatadogLogsArchiveGCS_basic(t *testing.T) {
	t.Parallel()
	if !isReplaying() {
		t.Skip("This test only supports replaying")
	}
	ctx, accProviders := testAccProviders(context.Background(), t)
	client := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckArchiveAndIntegrationGCSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveGCSConfigForCreation(client),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(accProvider),
					resource.TestCheckNoResourceAttr("datadog_logs_archive.my_gcs_archive", "gcs"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "name", "my first gcs archive"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "query", "service:tata"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs_archive.0.bucket", "dd-logs-test-datadog-api-client-go"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs_archive.0.client_email", fmt.Sprintf("%s@awesome-project-id.iam.gserviceaccount.com", client)),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs_archive.0.project_id", client),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs_archive.0.path", "/path/blah"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "include_tags", "false"),
					resource.TestCheckNoResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "rehydration_max_scan_size_in_gb"),
				),
			},
		},
	})
}

// create: Ok s3
func archiveS3ConfigForCreation(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  account_id         = "%s"
  role_name          = "testacc-datadog-integration-role"
}

resource "datadog_logs_archive" "my_s3_archive" {
  depends_on = ["datadog_integration_aws.account"]
  name = "my first s3 archive"
  query = "service:tutu"
  s3_archive {
    bucket 		 = "my-bucket"
    path 		 = "/path/foo"
    account_id   = "%s"
    role_name    = "testacc-datadog-integration-role"
  }
  rehydration_tags = ["team:intake", "team:app"]
  include_tags = true
	rehydration_max_scan_size_in_gb = 123
}`, uniq, uniq)
}

func TestAccDatadogLogsArchiveS3_basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckArchiveAndIntegrationAWSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveS3ConfigForCreation(accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(accProvider),
					resource.TestCheckNoResourceAttr("datadog_logs_archive.my_s3_archive", "s3"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "name", "my first s3 archive"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "query", "service:tutu"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.bucket", "my-bucket"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.role_name", "testacc-datadog-integration-role"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.path", "/path/foo"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "rehydration_tags.0", "team:intake"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "rehydration_tags.1", "team:app"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "rehydration_max_scan_size_in_gb", "123"),
				),
			},
		},
	})
}

// update: OK
func archiveS3ConfigForUpdate(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  account_id = "%s"
  role_name  = "testacc-datadog-integration-role"
}

resource "datadog_logs_archive" "my_s3_archive" {
  depends_on = ["datadog_integration_aws.account"]
  name       = "my first s3 archive after update"
  query      = "service:tutu"
  s3_archive {
  	bucket 		 = "my-bucket"
	path 		 = "/path/foo"
	account_id   = "%s"
	role_name    = "testacc-datadog-integration-role"
  }
  include_tags = false
	rehydration_max_scan_size_in_gb = 345
}`, uniq, uniq)
}

func TestAccDatadogLogsArchiveS3Update_basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	accProvider := testAccProvider(t, accProviders)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckArchiveAndIntegrationAWSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveS3ConfigForCreation(accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "name", "my first s3 archive"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.bucket", "my-bucket"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.role_name", "testacc-datadog-integration-role"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.path", "/path/foo"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "rehydration_max_scan_size_in_gb", "123"),
				),
			},
			{
				Config: archiveS3ConfigForUpdate(accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "name", "my first s3 archive after update"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "rehydration_tags.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "include_tags", "false"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.bucket", "my-bucket"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.role_name", "testacc-datadog-integration-role"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3_archive.0.path", "/path/foo"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "rehydration_max_scan_size_in_gb", "345"),
				),
			},
		},
	})
}

func testAccCheckArchiveExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := archiveExistsChecker(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func archiveExistsChecker(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_archive" {
			id := r.Primary.ID
			if _, _, err := apiInstances.GetLogsArchivesApiV2().GetLogsArchive(ctx, id); err != nil {
				return fmt.Errorf("received an error when retrieving archive, (%s)", err)
			}
		}
	}
	return nil
}

func testAccCheckArchiveAndIntegrationAzureDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		err := testAccCheckArchiveDestroy(accProvider)(s)
		if err != nil {
			return err
		}
		err = checkIntegrationAzureDestroy(accProvider)(s)
		return err
	}
}

func testAccCheckArchiveAndIntegrationGCSDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		err := testAccCheckArchiveDestroy(accProvider)(s)
		if err != nil {
			return err
		}
		err = checkIntegrationGCPDestroy(accProvider)(s)
		return err
	}
}

func testAccCheckArchiveAndIntegrationAWSDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		err := testAccCheckArchiveDestroy(accProvider)(s)
		if err != nil {
			return err
		}
		err = checkIntegrationAWSDestroy(accProvider)(s)
		return err
	}
}

func testAccCheckArchiveDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth
		if err := archiveDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func archiveDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_archive" {
			id := r.Primary.ID
			err := utils.Retry(2, 5, func() error {
				if r.Primary.ID != "" {
					_, httpresp, err := apiInstances.GetLogsArchivesApiV2().GetLogsArchive(ctx, id)
					if err != nil {
						if httpresp != nil && httpresp.StatusCode == 404 {
							return nil
						}
						return &utils.FatalError{Prob: fmt.Sprintf("received an error retrieving logs archives %s", err)}
					}
					return &utils.RetryableError{Prob: "logs archive still exists"}
				}
				return nil
			})
			return err
		}
	}
	return nil
}
