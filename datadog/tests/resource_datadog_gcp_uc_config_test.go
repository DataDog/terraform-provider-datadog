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

func TestAccGcpUcConfigBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogGcpUcConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogGcpUcConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGcpUcConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "billing_account_id", "123456_ABCDEF_123456"),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "bucket_name", "test-gcp-bucket"),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "export_dataset_name", "billing"),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "export_prefix", "datadog_cloud_cost_detailed_usage_export"),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "export_project_name", "test-gcp-project"),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "service_account", "test-service-account@test-project.iam.gserviceaccount.com"),
				),
			},
			{
				Config: testAccCheckDatadogGcpUcConfigDataSource(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGcpUcConfigExists(providers.frameworkProvider),
					// Check resource attributes
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "billing_account_id", "123456_ABCDEF_123456"),
					// Check data source attributes
					resource.TestCheckResourceAttrPair(
						"datadog_gcp_uc_config.foo", "bucket_name",
						"data.datadog_gcp_uc_config.bar", "bucket_name"),
					resource.TestCheckResourceAttrPair(
						"datadog_gcp_uc_config.foo", "export_project_name",
						"data.datadog_gcp_uc_config.bar", "export_project_name"),
					resource.TestCheckResourceAttr(
						"data.datadog_gcp_uc_config.bar", "status", "active"),
				),
			},
		},
	})
}

func TestAccGcpUcConfigImport(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogGcpUcConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogGcpUcConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGcpUcConfigExists(providers.frameworkProvider),
				),
			},
			{
				ResourceName:      "datadog_gcp_uc_config.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogGcpUcConfigBasic(uniq string) string {
	return `resource "datadog_gcp_uc_config" "foo" {
    billing_account_id = "123456_ABCDEF_123456"
    bucket_name = "test-gcp-bucket"
    export_dataset_name = "billing"
    export_prefix = "datadog_cloud_cost_detailed_usage_export"
    export_project_name = "test-gcp-project"
    service_account = "test-service-account@test-project.iam.gserviceaccount.com"
}`
}

func testAccCheckDatadogGcpUcConfigDataSource(uniq string) string {
	return `resource "datadog_gcp_uc_config" "foo" {
    billing_account_id = "123456_ABCDEF_123456"
    bucket_name = "test-gcp-bucket"
    export_dataset_name = "billing"
    export_prefix = "datadog_cloud_cost_detailed_usage_export"
    export_project_name = "test-gcp-project"
    service_account = "test-service-account@test-project.iam.gserviceaccount.com"
}

data "datadog_gcp_uc_config" "bar" {
    cloud_account_id = datadog_gcp_uc_config.foo.id
}`
}

func testAccCheckDatadogGcpUcConfig(uniq string) string {
	// Deprecated - kept for backwards compatibility
	return testAccCheckDatadogGcpUcConfigBasic(uniq)
}

func testAccCheckDatadogGcpUcConfigDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := GcpUcConfigDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func GcpUcConfigDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_gcp_uc_config" {
				continue
			}

			cloudAccountId, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
			resp, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetCostGCPUsageCostConfig(auth, cloudAccountId)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving GcpUcConfig %s", err)}
			}
			// Check if resource is archived (deleted)
			responseData := resp.GetData()
			if attributes, ok := responseData.GetAttributesOk(); ok {
				if attributes.GetStatus() == "archived" {
					return nil // Resource is properly deleted (archived)
				}
			}
			return &utils.RetryableError{Prob: fmt.Sprintf("GcpUcConfig still exists with status other than archived")}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogGcpUcConfigExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := gcpUcConfigExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func gcpUcConfigExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_gcp_uc_config" {
			continue
		}

		cloudAccountId, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
		_, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetCostGCPUsageCostConfig(auth, cloudAccountId)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving GcpUcConfig")
		}
	}
	return nil
}
