package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

// Adding a CUD metadata export to an account already managed in Terraform is
// purely additive. If cud_metadata_config ever regains RequiresReplace, this
// plan becomes a destroy-and-create, which costs the customer an ingestion gap
// and a full revalidation. The plan check is the guard against that.
func TestAccGcpUcConfigCudMetadataAddedInPlace(t *testing.T) {
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
					resource.TestCheckNoResourceAttr(
						"datadog_gcp_uc_config.foo", "cud_metadata_config.dataset_id"),
				),
			},
			{
				Config: testAccCheckDatadogGcpUcConfigWithCudMetadata(uniq),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"datadog_gcp_uc_config.foo", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGcpUcConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "cud_metadata_config.project_id", "test-gcp-project"),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "cud_metadata_config.dataset_id", "committed_usage_discounts"),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "cud_metadata_config.table_id", "cud_subscriptions_export"),
				),
			},
		},
	})
}

// A customer can configure their CUD metadata export in the Datadog UI, leaving
// no block in their configuration. Optional + Computed has to read that value
// back and leave it alone. Without Computed, Terraform reads a value the config
// does not declare and proposes deleting it, so the next apply silently reverts
// the customer's setup.
//
// ExpectNonEmptyPlan is false (the default, stated here because it *is* the
// assertion): the framework plans after the step and fails on any diff.
func TestAccGcpUcConfigCudMetadataConfiguredOutOfBand(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	var cloudAccountID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogGcpUcConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogGcpUcConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGcpUcConfigExists(providers.frameworkProvider),
					func(s *terraform.State) error {
						res, ok := s.RootModule().Resources["datadog_gcp_uc_config.foo"]
						if !ok {
							return fmt.Errorf("datadog_gcp_uc_config.foo not found in state")
						}
						cloudAccountID = res.Primary.ID
						return nil
					},
				),
			},
			{
				// Configure the export behind Terraform's back, the way the
				// Datadog UI would.
				PreConfig: func() {
					if err := setGcpCudMetadataConfigOutOfBand(providers.frameworkProvider, cloudAccountID); err != nil {
						t.Fatalf("failed to set CUD metadata config out of band: %s", err)
					}
				},
				// Unchanged, and deliberately still carries no cud_metadata_config.
				Config:             testAccCheckDatadogGcpUcConfigBasic(uniq),
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					// Read back into state rather than ignored.
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "cud_metadata_config.dataset_id", "committed_usage_discounts"),
				),
			},
		},
	})
}

// Optional + Computed cannot distinguish "the customer removed this block" from
// "the customer never wrote one", so removing it is a no-op rather than an
// opt-out. That is the accepted trade for not clobbering UI-configured accounts
// above. Pinned here so it reads as intended behaviour and not a bug.
func TestAccGcpUcConfigCudMetadataBlockRemovalIsNoop(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogGcpUcConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogGcpUcConfigWithCudMetadata(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogGcpUcConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "cud_metadata_config.dataset_id", "committed_usage_discounts"),
				),
			},
			{
				Config: testAccCheckDatadogGcpUcConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					// Still configured. To opt out, PATCH an empty config or use
					// the Datadog UI.
					resource.TestCheckResourceAttr(
						"datadog_gcp_uc_config.foo", "cud_metadata_config.dataset_id", "committed_usage_discounts"),
				),
			},
		},
	})
}

// setGcpCudMetadataConfigOutOfBand PATCHes a CUD metadata config directly,
// bypassing Terraform, to simulate a customer configuring it in the Datadog UI.
func setGcpCudMetadataConfigOutOfBand(accProvider *fwprovider.FrameworkProvider, cloudAccountID string) error {
	id, err := strconv.ParseInt(cloudAccountID, 10, 64)
	if err != nil {
		return err
	}

	cudMetadataConfig := datadogV2.NewGCPUsageCostConfigCudMetadataConfigWithDefaults()
	cudMetadataConfig.SetProjectId("test-gcp-project")
	cudMetadataConfig.SetDatasetId("committed_usage_discounts")
	cudMetadataConfig.SetTableId("cud_subscriptions_export")

	attributes := datadogV2.NewGCPUsageCostConfigPatchRequestAttributesWithDefaults()
	attributes.SetCudMetadataConfig(*cudMetadataConfig)

	body := datadogV2.NewGCPUsageCostConfigPatchRequestWithDefaults()
	body.Data = *datadogV2.NewGCPUsageCostConfigPatchDataWithDefaults()
	body.Data.SetAttributes(*attributes)

	_, _, err = accProvider.DatadogApiInstances.GetCloudCostManagementApiV2().
		UpdateCostGCPUsageCostConfig(accProvider.Auth, id, *body)
	return err
}

func testAccCheckDatadogGcpUcConfigWithCudMetadata(uniq string) string {
	return `resource "datadog_gcp_uc_config" "foo" {
    billing_account_id = "123456_ABCDEF_123456"
    bucket_name = "test-gcp-bucket"
    export_dataset_name = "billing"
    export_prefix = "datadog_cloud_cost_detailed_usage_export"
    export_project_name = "test-gcp-project"
    service_account = "test-service-account@test-project.iam.gserviceaccount.com"

    cud_metadata_config {
        project_id = "test-gcp-project"
        dataset_id = "committed_usage_discounts"
        table_id   = "cud_subscriptions_export"
    }
}`
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
