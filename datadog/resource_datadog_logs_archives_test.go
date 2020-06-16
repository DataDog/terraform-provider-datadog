package datadog

import (
	"context"
	"fmt"
	"testing"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"gopkg.in/h2non/gock.v1"
	"net/http"
)

const archiveAzureConfigForCreation = `
resource "datadog_logs_archive" "my_azure_archive" {
	name = "my first azure archive"
	query = "service:toto"
	azure = {
		container 		= "container"
		client_id 		= "clientId"
		tenant_id       = "tenantId"
		storage_account = "storageAccount"
	}
}
`

var archiveAzure = datadogV2.LogsArchiveCreateRequest{
	Data: &datadogV2.LogsArchiveCreateRequestDefinition{
		Attributes: &datadogV2.LogsArchiveCreateRequestAttributes{
			Destination: datadogV2.LogsArchiveCreateRequestDestination{
				LogsArchiveDestinationAzure: &datadogV2.LogsArchiveDestinationAzure{
					Container: "my-container",
					Integration: datadogV2.LogsArchiveIntegrationAzure{
						ClientId: "aaaaaaaa-1a1a-1a1a-1a1a-aaaaaaaaaaaa",
						TenantId: "aaaaaaaa-1a1a-1a1a-1a1a-aaaaaaaaaaaa",
					},
					Path:           datadogV2.PtrString("/path/blou"),
					Region:         datadogV2.PtrString("my-region"),
					StorageAccount: "storageAccount",
					Type:           "azure",
				},
			},
			Name:  "datadog-api-client-go Tests Archive",
			Query: "service:toto",
		},
		Type: "archives",
	},
}

const archiveGCSConfigForCreation = `
resource "datadog_logs_archive" "my_gcs_archive" {
	name = "my first gcs archive"
	query = "service:toto"
	gcs = {
        bucket 		 = "bucket"
        path 	     = "/path/hello"
        client_email = "clientEmail"
        project_id   = "projectId"
	}
}
`

const archiveS3ConfigForCreation = `
resource "datadog_logs_archive" "my_s3_archive" {
	name = "my first azure archive"
	query = "service:toto"
	s3 = {
        bucket 		 = "bucket"
        path 		 = "/path/hello"
        client_email = "clientEmail"
        project_id   = "projectId"
        account_id   = "accountId"
        role_name    = "roleName"
	}
}
`

//Test
// create: OK azure

func TestAccDatadogLogsArchive_basic(t *testing.T) {
	outputArchiveStr := ""
	gock.New("https://api.datadoghq.com").Post("/api/v2/logs/config/archives").MatchType("json").JSON(archiveAzure).Reply(200).Type("json").BodyString(outputArchiveStr)
	accProviders := testAccProvidersWithHttpClient(t, http.DefaultClient)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckArchiveDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
				},
				Config: archiveAzureConfigForCreation,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_archive.my_archive_test", "name", "my first azure archive"),
				),
			},
		},
	})
	fmt.Printf("Finished !")
}

// create: Ok s3
// create: Ok gcs
// create: type azure + azure, s3 defined => Fail
// create: type azure + gcs defined => Fail
// create: type unknown => Fail
// update: OK
// update: does not exist
// delete: OK
// delete: does not exist

func testAccCheckArchiveExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if err := archiveExistsChecker(s, authV2, datadogClientV2); err != nil {
			return err
		}
		return nil
	}
}

func archiveExistsChecker(s *terraform.State, authV2 context.Context, datadogClientV2 *datadogV2.APIClient) error {
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

func testAccCheckArchiveDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if err := archiveDestroyHelper(s, authV2, datadogClientV2); err != nil {
			return err
		}
		return nil
	}
}

func archiveDestroyHelper(s *terraform.State, authV2 context.Context, datadogClientV2 *datadogV2.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_archive" {
			id := r.Primary.ID
			archive, httpresp, err := datadogClientV2.LogsArchivesApi.GetLogsArchive(authV2, id).Execute()
			if err != nil {
				if httpresp.StatusCode == 404 {
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
