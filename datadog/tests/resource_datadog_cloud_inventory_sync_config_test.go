package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const syncConfigByIDPath = "/api/v2/cloudinventoryservice/syncconfigs/%s"

func TestAccCloudInventorySyncConfigBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	_ = uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogCloudInventorySyncConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudInventorySyncConfigAWS(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogCloudInventorySyncConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.foo", "cloud_provider", "aws"),
					resource.TestCheckResourceAttr("datadog_cloud_inventory_sync_config.foo", "aws.aws_account_id", "123456789012"),
				),
			},
		},
	})
}

func testAccCheckDatadogCloudInventorySyncConfigAWS() string {
	return `
resource "datadog_cloud_inventory_sync_config" "foo" {
    cloud_provider = "aws"
    
    aws {
        aws_account_id            = "123456789012"
        destination_bucket_name   = "test-inventory-bucket"
        destination_bucket_region = "us-east-1"
        destination_prefix        = "inventory/"
    }
}`
}

func testAccCheckDatadogCloudInventorySyncConfigDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := cloudInventorySyncConfigDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func cloudInventorySyncConfigDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_cloud_inventory_sync_config" {
				continue
			}
			id := r.Primary.ID
			path := fmt.Sprintf(syncConfigByIDPath, id)

			_, httpResp, err := utils.SendRequest(auth, apiInstances.HttpClient, http.MethodGet, path, nil)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving CloudInventorySyncConfig %s", err)}
			}
			return &utils.RetryableError{Prob: "CloudInventorySyncConfig still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogCloudInventorySyncConfigExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := cloudInventorySyncConfigExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func cloudInventorySyncConfigExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_cloud_inventory_sync_config" {
			continue
		}
		id := r.Primary.ID
		path := fmt.Sprintf(syncConfigByIDPath, id)

		_, httpResp, err := utils.SendRequest(auth, apiInstances.HttpClient, http.MethodGet, path, nil)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving CloudInventorySyncConfig")
		}
	}
	return nil
}
