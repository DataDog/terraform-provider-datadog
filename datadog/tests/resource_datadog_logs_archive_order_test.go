package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

func ArchiveOrderConfig() string {
	return `
resource "datadog_logs_archive_order" "archives" {
	archive_ids = []
}`
}

func ArchiveOrderEmptyConfig() string {
	return `
resource "datadog_logs_archive_order" "archives" {
}`
}

func TestAccDatadogLogsArchiveOrder_basic(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckPipelineDestroy(accProvider),
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
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckPipelineDestroy(accProvider),
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

func testAccCheckArchiveOrderExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := archiveOrderExistsChecker(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func archiveOrderExistsChecker(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_archive_order" {
			if _, _, err := apiInstances.GetLogsArchivesApiV2().GetLogsArchiveOrder(ctx); err != nil {
				return fmt.Errorf("received an error when retrieving archive order, (%s)", err)
			}
		}
	}
	return nil
}

func testAccCheckArchiveOrderResourceMatch(accProvider func() (*schema.Provider, error), name string, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		resourceType := strings.Split(name, ".")[0]
		elemNo, _ := strconv.Atoi(strings.Split(key, ".")[1])
		for _, r := range s.RootModule().Resources {
			if r.Type == resourceType {
				archiveOrder, _, err := apiInstances.GetLogsArchivesApiV2().GetLogsArchiveOrder(auth)
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
