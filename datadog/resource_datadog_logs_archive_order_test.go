package datadog

import (
	"context"
	"fmt"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func ArchiveOrderConfig() string {
	return fmt.Sprintf(`
resource "datadog_logs_archive_order" "archives" {
	archive_ids = []
}`)
}

func ArchiveOrderEmptyConfig() string {
	return fmt.Sprintf(`
resource "datadog_logs_archive_order" "archives" {
}`)
}

func TestAccDatadogLogsArchiveOrder_basic(t *testing.T) {
	rec := initRecorder(t)
	accProviders, _, cleanup := testAccProviders(t, rec)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckPipelineDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: ArchiveOrderConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveOrderExists(accProvider),
					resource.TestCheckResourceAttrSet(
						"datadog_logs_archive_order.archives", "archive_ids.#"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDatadogLogsArchiveOrder_empty(t *testing.T) {
	rec := initRecorder(t)
	accProviders, _, cleanup := testAccProviders(t, rec)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckPipelineDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: ArchiveOrderEmptyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveOrderExists(accProvider),
					resource.TestCheckResourceAttrSet(
						"datadog_logs_archive_order.archives", "archive_ids.#"),
				),
			},
		},
	})
}

func testAccCheckArchiveOrderExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV1 := providerConf.AuthV1

		if err := ArchiveOrderExistsChecker(s, authV1, datadogClientV2); err != nil {
			return err
		}
		return nil
	}
}

func ArchiveOrderExistsChecker(s *terraform.State, authV1 context.Context, datadogClientV1 *datadogV2.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_archive_order" {
			if _, _, err := datadogClientV1.LogsArchivesApi.GetLogsArchiveOrder(authV1).Execute(); err != nil {
				return fmt.Errorf("received an error when retrieving archive order, (%s)", err)
			}
		}
	}
	return nil
}
