package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogIntegrationAwsTagFilter_Basic(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueID := uniqueAWSAccountID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationAwsTagFilterDestroy(accProvider, "datadog_integration_aws_tag_filter.testing_aws_tag_filter"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAwsTagFilterBasic(uniqueID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsTagFilterExists(accProvider, "datadog_integration_aws_tag_filter.testing_aws_tag_filter"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_tag_filter.testing_aws_tag_filter", "account_id", uniqueID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_tag_filter.testing_aws_tag_filter", "namespace", "application_elb"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_tag_filter.testing_aws_tag_filter", "tag_filter_str", "test:filter"),
				),
			},
		},
	})
}

func TestAccDatadogIntegrationAwsTagFilter_BasicAccessKey(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	accessKeyID := uniqueAWSAccessKeyID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationAwsTagFilterDestroy(accProvider, "datadog_integration_aws_tag_filter.testing_aws_tag_filter"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAwsTagFilterBasicAccessKey(accessKeyID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsTagFilterExists(accProvider, "datadog_integration_aws_tag_filter.testing_aws_tag_filter"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_tag_filter.testing_aws_tag_filter", "account_id", accessKeyID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_tag_filter.testing_aws_tag_filter", "namespace", "application_elb"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_tag_filter.testing_aws_tag_filter", "tag_filter_str", "test:filter"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationAwsTagFilterBasic(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  account_id                       = "%s"
  role_name                        = "testacc-datadog-integration-role"
}

resource "datadog_integration_aws_tag_filter" "testing_aws_tag_filter" {
	account_id     = datadog_integration_aws.account.account_id
	namespace      = "application_elb"
	tag_filter_str = "test:filter"
    depends_on     = [datadog_integration_aws.account]
}`, uniq)
}

func testAccCheckDatadogIntegrationAwsTagFilterBasicAccessKey(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
	access_key_id     = "%s"
	secret_access_key = "testacc-datadog-integration-secret"
}

resource "datadog_integration_aws_tag_filter" "testing_aws_tag_filter" {
	account_id     = datadog_integration_aws.account.access_key_id
	namespace      = "application_elb"
	tag_filter_str = "test:filter"
	depends_on     = [datadog_integration_aws.account]
}`, uniq)
}

func testAccCheckDatadogIntegrationAwsTagFilterExists(accProvider func() (*schema.Provider, error), resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceID := s.RootModule().Resources[resourceName].Primary.ID
		_, tfNamespace, _ := utils.AccountAndNamespaceFromID(resourceID)
		namespace := datadogV1.AWSNamespace(tfNamespace)

		filters, err := listFiltersHelper(accProvider, resourceID)
		if err != nil {
			return err
		}

		for _, filter := range *filters {
			if filter.GetNamespace() == namespace {
				if len(filter.GetTagFilterStr()) == 0 {
					return fmt.Errorf("tag_filter_str is empty for resource %s", namespace)
				}
				return nil
			}
		}

		return nil
	}
}

func testAccCheckDatadogIntegrationAwsTagFilterDestroy(accProvider func() (*schema.Provider, error), resourceName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		resourceID := s.RootModule().Resources[resourceName].Primary.ID
		_, tfNamespace, _ := utils.AccountAndNamespaceFromID(resourceID)
		namespace := datadogV1.AWSNamespace(tfNamespace)

		filters, err := listFiltersHelper(accProvider, resourceID)
		if err != nil {
			errObj := err.(datadogV1.GenericOpenAPIError)
			if matched, _ := regexp.MatchString("AWS account [0-9]+ does not exist in integration", string(errObj.Body())); matched {
				return nil
			}
			return err
		}

		for _, filter := range *filters {
			if filter.GetNamespace() == namespace {
				if len(filter.GetTagFilterStr()) != 0 {
					return fmt.Errorf("tag_filter_str is not empty for namespace %s", namespace)
				}
				return nil
			}
		}

		return nil
	}
}

func listFiltersHelper(accProvider func() (*schema.Provider, error), resourceID string) (*[]datadogV1.AWSTagFilter, error) {
	provider, _ := accProvider()
	providerConf := provider.Meta().(*datadog.ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	filters := []datadogV1.AWSTagFilter{}
	accountID, _, err := utils.AccountAndNamespaceFromID(resourceID)
	if err != nil {
		return nil, err
	}

	resp, _, err := datadogClient.AWSIntegrationApi.ListAWSTagFilters(auth, accountID)
	if err != nil {
		return nil, err
	}
	filters = append(filters, resp.GetFilters()...)

	return &filters, nil
}
