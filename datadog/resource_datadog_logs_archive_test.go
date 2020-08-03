package datadog

import (
	"context"
	"fmt"
	"testing"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

//Test
// create: OK azure
func archiveAzureConfigForCreation(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_azure" "an_azure_integration" {
  tenant_name   = "%s"
  client_id     = "testc7f6-1234-5678-9101-3fcbf464test"
  client_secret = "testingx./Sw*g/Y33t..R1cH+hScMDt"
}

resource "datadog_logs_archive" "my_azure_archive" {
  depends_on = ["datadog_integration_azure.an_azure_integration"]
  name  = "my first azure archive"
  query = "service:toto"
  azure = {
    container 		= "my-container"
    tenant_id 		= "%s"
    client_id       = "testc7f6-1234-5678-9101-3fcbf464test"
    storage_account = "storageAccount"
    path            = "/path/blou"
  }
}
`, uniq, uniq)
}

func TestAccDatadogLogsArchiveAzure_basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	tenantName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckArchiveAndIntegrationAzureDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveAzureConfigForCreation(tenantName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "name", "my first azure archive"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "query", "service:toto"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure.container", "my-container"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure.client_id", "testc7f6-1234-5678-9101-3fcbf464test"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure.tenant_id", tenantName),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure.storage_account", "storageAccount"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure.path", "/path/blou"),
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
  gcs        = {
    bucket 		 = "dd-logs-test-datadog-api-client-go"
	path 	     = "/path/blah"
	client_email = "%s@awesome-project-id.iam.gserviceaccount.com"
	project_id   = "%s"
  }
}`, uniq, uniq, uniq, uniq)
}

func TestAccDatadogLogsArchiveGCS_basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	client := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckArchiveAndIntegrationGCSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveGCSConfigForCreation(client),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "name", "my first gcs archive"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "query", "service:tata"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs.bucket", "dd-logs-test-datadog-api-client-go"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs.client_email", fmt.Sprintf("%s@awesome-project-id.iam.gserviceaccount.com", client)),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs.project_id", client),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs.path", "/path/blah"),
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
  s3 = {
    bucket 		 = "my-bucket"
    path 		 = "/path/foo"
    account_id   = "%s"
    role_name    = "testacc-datadog-integration-role"
  }
}`, uniq, uniq)
}

func TestAccDatadogLogsArchiveS3_basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	accountID := uniqueAWSAccountID(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckArchiveAndIntegrationAWSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveS3ConfigForCreation(accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "name", "my first s3 archive"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "query", "service:tutu"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3.bucket", "my-bucket"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3.account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3.role_name", "testacc-datadog-integration-role"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3.path", "/path/foo"),
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
  s3 = {
  	bucket 		 = "my-bucket"
	path 		 = "/path/foo"
	account_id   = "%s"
	role_name    = "testacc-datadog-integration-role"
  }
}`, uniq, uniq)
}

func TestAccDatadogLogsArchiveS3Update_basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	accountID := uniqueAWSAccountID(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckArchiveAndIntegrationAWSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveS3ConfigForCreation(accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "name", "my first s3 archive"),
				),
			},
			{
				Config: archiveS3ConfigForUpdate(accountID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "name", "my first s3 archive after update"),
				),
			},
		},
	})
}

func testAccCheckArchiveExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if err := archiveExistsChecker(authV2, s, datadogClientV2); err != nil {
			return err
		}
		return nil
	}
}

func archiveExistsChecker(authV2 context.Context, s *terraform.State, datadogClientV2 *datadogV2.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_archive" {
			id := r.Primary.ID
			if _, _, err := datadogClientV2.LogsArchivesApi.GetLogsArchive(authV2, id).Execute(); err != nil {
				return fmt.Errorf("received an error when retrieving archive, (%s)", err)
			}
		}
	}
	return nil
}

func testAccCheckArchiveAndIntegrationAzureDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		err := testAccCheckArchiveDestroy(accProvider)(s)
		if err != nil {
			return err
		}
		err = checkIntegrationAzureDestroy(accProvider)(s)
		return err
	}
}

func testAccCheckArchiveAndIntegrationGCSDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		err := testAccCheckArchiveDestroy(accProvider)(s)
		if err != nil {
			return err
		}
		err = checkIntegrationGCPDestroy(accProvider)(s)
		return err
	}
}

func testAccCheckArchiveAndIntegrationAWSDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		err := testAccCheckArchiveDestroy(accProvider)(s)
		if err != nil {
			return err
		}
		err = checkIntegrationAWSDestroy(accProvider)(s)
		return err
	}
}

func testAccCheckArchiveDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2
		if err := archiveDestroyHelper(authV2, s, datadogClientV2); err != nil {
			return err
		}
		return nil
	}
}

func archiveDestroyHelper(authV2 context.Context, s *terraform.State, datadogClientV2 *datadogV2.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_archive" {
			id := r.Primary.ID
			archive, httpresp, err := datadogClientV2.LogsArchivesApi.GetLogsArchive(authV2, id).Execute()
			if err != nil {
				if httpresp != nil && httpresp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error when retrieving pipeline, (%s)", err)
			}
			if &archive != nil {
				return fmt.Errorf("archive still exists")
			}
		}

	}
	return nil
}
