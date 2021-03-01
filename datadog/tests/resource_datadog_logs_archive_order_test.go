package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
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
					testAccCheckArchiveOrderResourceMatch(accProvider, "datadog_logs_archive_order.archives", "archive_ids.0"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDatadogLogsArchiveOrder_empty(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
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
					testAccCheckArchiveOrderResourceMatch(accProvider, "datadog_logs_archive_order.archives", "archive_ids.0"),
				),
			},
		},
	})
}

func testAccCheckArchiveOrderExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV1 := providerConf.AuthV1

		if err := archiveOrderExistsChecker(authV1, s, datadogClientV2); err != nil {
			return err
		}
		return nil
	}
}

func archiveOrderExistsChecker(ctx context.Context, s *terraform.State, datadogClientV2 *datadogV2.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_archive_order" {
			if _, _, err := datadogClientV2.LogsArchivesApi.GetLogsArchiveOrder(ctx).Execute(); err != nil {
				return fmt.Errorf("received an error when retrieving archive order, (%s)", err)
			}
		}
	}
	return nil
}

func testAccCheckArchiveOrderResourceMatch(accProvider *schema.Provider, name string, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV2 := providerConf.DatadogClientV2
		authV1 := providerConf.AuthV1

		resourceType := strings.Split(name, ".")[0]
		elemNo, _ := strconv.Atoi(strings.Split(key, ".")[1])
		for _, r := range s.RootModule().Resources {
			if r.Type == resourceType {
				archiveOrder, _, err := datadogClientV2.LogsArchivesApi.GetLogsArchiveOrder(authV1).Execute()
				if err != nil {
					return fmt.Errorf("received an error when retrieving archive order, (%s)", err)
				}
				archiveIds := archiveOrder.Data.Attributes.GetArchiveIds()
				if elemNo >= len(archiveIds) {
					println("can't match (%s), there are only %s archives", key, strconv.Itoa(len(archiveIds)))
					return nil
				}
				return resource.TestCheckResourceAttr(name, key, archiveIds[elemNo])(s)
			}
		}
		return fmt.Errorf("didn't find %s", name)
	}
}
