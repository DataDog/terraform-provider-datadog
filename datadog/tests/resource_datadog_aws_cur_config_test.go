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

func TestAccAwsCurConfigBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAwsCurConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAwsCurConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAwsCurConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "account_id", "123456789012"),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "bucket_name", "test-cur-bucket"),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "bucket_region", "us-east-1"),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "report_name", "test-cur-report"),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "report_prefix", "test-cur-prefix"),
				),
			},
			{
				Config: testAccCheckDatadogAwsCurConfigDataSource(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAwsCurConfigExists(providers.frameworkProvider),
					// Check resource attributes
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "account_id", "123456789012"),
					// Check data source attributes
					resource.TestCheckResourceAttr(
						"data.datadog_aws_cur_config.bar", "account_id", "123456789012"),
					resource.TestCheckResourceAttrPair(
						"datadog_aws_cur_config.foo", "bucket_name",
						"data.datadog_aws_cur_config.bar", "bucket_name"),
					resource.TestCheckResourceAttrPair(
						"datadog_aws_cur_config.foo", "report_name",
						"data.datadog_aws_cur_config.bar", "report_name"),
					resource.TestCheckResourceAttr(
						"data.datadog_aws_cur_config.bar", "status", "active"),
				),
			},
		},
	})
}

func TestAccAwsCurConfigWithAccountFilters(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAwsCurConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAwsCurConfigWithFilters(uniq, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAwsCurConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "account_id", "123456789012"),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "account_filters.include_new_accounts", "true"),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "account_filters.excluded_accounts.0", "123456789012"),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "account_filters.included_accounts.#", "0"),
				),
			},
			{
				Config: testAccCheckDatadogAwsCurConfigWithFilters(uniq, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAwsCurConfigExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "account_filters.include_new_accounts", "false"),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "account_filters.included_accounts.0", "123456789013"),
					resource.TestCheckResourceAttr(
						"datadog_aws_cur_config.foo", "account_filters.excluded_accounts.#", "0"),
				),
			},
		},
	})
}

func TestAccAwsCurConfigImport(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAwsCurConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAwsCurConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAwsCurConfigExists(providers.frameworkProvider),
				),
			},
			{
				ResourceName:      "datadog_aws_cur_config.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogAwsCurConfigBasic(uniq string) string {
	return `resource "datadog_aws_cur_config" "foo" {
    account_id = "123456789012"
    bucket_name = "test-cur-bucket"
    bucket_region = "us-east-1"
    report_name = "test-cur-report"
    report_prefix = "test-cur-prefix"
}`
}

func testAccCheckDatadogAwsCurConfigWithFilters(uniq string, includeNewAccounts bool) string {
	if includeNewAccounts {
		// When include_new_accounts = true, use excluded_accounts
		return `resource "datadog_aws_cur_config" "foo" {
    account_id = "123456789012"
    bucket_name = "test-cur-bucket"
    bucket_region = "us-east-1"
    report_name = "test-cur-report"
    report_prefix = "test-cur-prefix"
    
    account_filters {
        include_new_accounts = true
        excluded_accounts = ["123456789012"]
    }
}`
	} else {
		// When include_new_accounts = false, use included_accounts
		return `resource "datadog_aws_cur_config" "foo" {
    account_id = "123456789012"
    bucket_name = "test-cur-bucket"
    bucket_region = "us-east-1"
    report_name = "test-cur-report"
    report_prefix = "test-cur-prefix"
    
    account_filters {
        include_new_accounts = false
        included_accounts = ["123456789013"]
    }
}`
	}
}

func testAccCheckDatadogAwsCurConfigDataSource(uniq string) string {
	return `resource "datadog_aws_cur_config" "foo" {
    account_id = "123456789012"
    bucket_name = "test-cur-bucket"
    bucket_region = "us-east-1"
    report_name = "test-cur-report"
    report_prefix = "test-cur-prefix"
}

data "datadog_aws_cur_config" "bar" {
    cloud_account_id = datadog_aws_cur_config.foo.id
}`
}

func testAccCheckDatadogAwsCurConfig(uniq string) string {
	// Deprecated - kept for backwards compatibility
	return testAccCheckDatadogAwsCurConfigBasic(uniq)
}

func testAccCheckDatadogAwsCurConfigDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := AwsCurConfigDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func AwsCurConfigDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_aws_cur_config" {
				continue
			}

			cloudAccountId, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
			resp, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetCostAWSCURConfig(auth, cloudAccountId)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AwsCurConfig %s", err)}
			}
			// Check if resource is archived (deleted)
			responseData := resp.GetData()
			if attributes, ok := responseData.GetAttributesOk(); ok {
				status := attributes.GetStatus()
				if status == "archived" {
					return nil // Resource is properly deleted (archived)
				}
			}
			return &utils.RetryableError{Prob: fmt.Sprintf("AwsCurConfig still exists with status other than archived")}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogAwsCurConfigExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := awsCurConfigExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func awsCurConfigExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_aws_cur_config" {
			continue
		}

		cloudAccountId, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
		_, httpResp, err := apiInstances.GetCloudCostManagementApiV2().GetCostAWSCURConfig(auth, cloudAccountId)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving AwsCurConfig")
		}
	}
	return nil
}
