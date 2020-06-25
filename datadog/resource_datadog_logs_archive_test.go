package datadog

import (
	"context"
	"fmt"
	"testing"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"os"
)

//Test
// create: OK azure

const archiveAzureConfigForCreation = `
resource "datadog_logs_archive" "my_azure_archive" {
  name  = "my first azure archive"
  query = "service:toto"
  azure = {
    container 		= "my-container"
    tenant_id 		= "my-tenant-id"
    client_id       = "testc7f6-1234-5678-9101-3fcbf464test"
    storage_account = "storageAccount"
    path            = "/path/blou"
  }
}
`

func getApiKey() string {
	if os.Getenv("DATADOG_API_KEY") != "" {
		return os.Getenv("DATADOG_API_KEY")
	}
	if os.Getenv("DD_API_KEY") != "" {
		return os.Getenv("DD_API_KEY")
	}
	return ""
}
func getAppKey() string {
	if os.Getenv("DATADOG_APP_KEY") != "" {
		return os.Getenv("DATADOG_APP_KEY")
	}
	if os.Getenv("DD_APP_KEY") != "" {
		return os.Getenv("DD_APP_KEY")
	}
	return ""
}

func TestAccDatadogLogsArchiveAzure_basic(t *testing.T) {
	rec := initRecorder(t)
	defer rec.Stop()
	httpClient := &http.Client{Transport: logging.NewTransport("Datadog", rec)}
	// At the moment there's no azure integration in tf so we manually:
	// 1. Create an api client with the right conf and the right recorder
	datadogClientV1 := buildDatadogClientV1(httpClient)
	authV1, err := buildAuthV1(getApiKey(), getAppKey(), "")
	if err != nil {
		t.Fatalf("Error creating Datadog Client context: %s", err)
	}
	var testAzureAcct = datadogV1.AzureAccount{
		ClientId:     datadogV1.PtrString("testc7f6-1234-5678-9101-3fcbf464test"),
		ClientSecret: datadogV1.PtrString("testingx./Sw*g/Y33t..R1cH+hScMDt"),
		TenantName:   datadogV1.PtrString("my-tenant-id"),
	}
	// 2. Create the azure account
	_, _, err = datadogClientV1.AzureIntegrationApi.CreateAzureIntegration(authV1).Body(testAzureAcct).Execute()
	if err != nil {
		t.Fatalf("Error creating Azure Account: Response %s: %v", err.(datadogV1.GenericOpenAPIError).Body(), err)
	}
	// 3. Destroy it at the end of the test
	defer deleteAzureIntegration(t, datadogClientV1, authV1, testAzureAcct)

	accProviders := testAccProvidersWithHttpClient(t, httpClient)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckArchiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveAzureConfigForCreation,
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
						"datadog_logs_archive.my_azure_archive", "azure.tenant_id", "my-tenant-id"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure.storage_account", "storageAccount"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_azure_archive", "azure.path", "/path/blou"),
				),
			},
		},
	})
}

func deleteAzureIntegration(t *testing.T, datadogClientV1 *datadogV1.APIClient, authV1 context.Context, azureAcct datadogV1.AzureAccount) {
	_, _, err := datadogClientV1.AzureIntegrationApi.DeleteAzureIntegration(authV1).Body(azureAcct).Execute()
	if err != nil {
		t.Fatalf("Error deleting Azure Account: Response %s: %v", err.(datadogV1.GenericOpenAPIError).Body(), err)
	}
}

// create: Ok gcs
const archiveGCSConfigForCreation = `

resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id     = "super-awesome-project-id"
  private_key_id = "1234567890123456789012345678901234567890"
  private_key    = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  client_email   = "awesome-service-account@awesome-project-id.iam.gserviceaccount.com"
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
	client_email = "awesome-service-account@awesome-project-id.iam.gserviceaccount.com"
	project_id   = "super-awesome-project-id"
  }
}
`

func TestAccDatadogLogsArchiveGCS_basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckArchiveAndIntegrationGCSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveGCSConfigForCreation,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "name", "my first gcs archive"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "query", "service:tata"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs.bucket", "dd-logs-test-datadog-api-client-go"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs.client_email", "awesome-service-account@awesome-project-id.iam.gserviceaccount.com"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs.project_id", "super-awesome-project-id"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_gcs_archive", "gcs.path", "/path/blah"),
				),
			},
		},
	})
}

// create: Ok s3
const archiveS3ConfigForCreation = `
resource "datadog_integration_aws" "account" {
  account_id         = "001234567888"
  role_name          = "testacc-datadog-integration-role"
}

resource "datadog_logs_archive" "my_s3_archive" {
  depends_on = ["datadog_integration_aws.account"]
  name = "my first s3 archive"
  query = "service:tutu"
  s3 = {
    bucket 		 = "my-bucket"
    path 		 = "/path/foo"
    account_id   = "001234567888"
    role_name    = "testacc-datadog-integration-role"
  }
}
`

func TestAccDatadogLogsArchiveS3_basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckArchiveAndIntegrationAWSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveS3ConfigForCreation,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "name", "my first s3 archive"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "query", "service:tutu"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3.bucket", "my-bucket"),
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "s3.account_id", "001234567888"),
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
const archiveS3ConfigForUpdate = `

resource "datadog_integration_aws" "account" {
  account_id = "001234567888"
  role_name  = "testacc-datadog-integration-role"
}

resource "datadog_logs_archive" "my_s3_archive" {
  depends_on = ["datadog_integration_aws.account"]
  name       = "my first s3 archive after update"
  query      = "service:tutu"
  s3 = {
  	bucket 		 = "my-bucket"
	path 		 = "/path/foo"
	account_id   = "001234567888"
	role_name    = "testacc-datadog-integration-role"
  }
}
`

func TestAccDatadogLogsArchiveS3Update_basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckArchiveAndIntegrationAWSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: archiveS3ConfigForCreation,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_s3_archive", "name", "my first s3 archive"),
				),
			},
			{
				Config: archiveS3ConfigForUpdate,
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
