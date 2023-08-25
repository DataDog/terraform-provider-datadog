package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	dd "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIntegrationAwsTagFilter_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqueID := uniqueAWSAccountID(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationAwsTagFilterDestroy(providers.frameworkProvider, "datadog_integration_aws_tag_filter.testing_aws_tag_filter"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAwsTagFilterBasic(uniqueID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsTagFilterExists(providers.frameworkProvider, "datadog_integration_aws_tag_filter.testing_aws_tag_filter"),
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
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accessKeyID := uniqueAWSAccessKeyID(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationAwsTagFilterDestroy(providers.frameworkProvider, "datadog_integration_aws_tag_filter.testing_aws_tag_filter"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAwsTagFilterBasicAccessKey(accessKeyID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsTagFilterExists(providers.frameworkProvider, "datadog_integration_aws_tag_filter.testing_aws_tag_filter"),
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

func testAccCheckDatadogIntegrationAwsTagFilterExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
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

func testAccCheckDatadogIntegrationAwsTagFilterDestroy(accProvider *fwprovider.FrameworkProvider, resourceName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		resourceID := s.RootModule().Resources[resourceName].Primary.ID
		_, tfNamespace, _ := utils.AccountAndNamespaceFromID(resourceID)
		namespace := datadogV1.AWSNamespace(tfNamespace)

		filters, err := listFiltersHelper(accProvider, resourceID)
		if err != nil {
			errObj := err.(dd.GenericOpenAPIError)
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

func listFiltersHelper(accProvider *fwprovider.FrameworkProvider, resourceID string) (*[]datadogV1.AWSTagFilter, error) {
	apiInstances := accProvider.DatadogApiInstances
	auth := accProvider.Auth

	filters := []datadogV1.AWSTagFilter{}
	accountID, _, err := utils.AccountAndNamespaceFromID(resourceID)
	if err != nil {
		return nil, err
	}

	resp, _, err := apiInstances.GetAWSIntegrationApiV1().ListAWSTagFilters(auth, accountID)
	if err != nil {
		return nil, err
	}
	filters = append(filters, resp.GetFilters()...)

	return &filters, nil
}
