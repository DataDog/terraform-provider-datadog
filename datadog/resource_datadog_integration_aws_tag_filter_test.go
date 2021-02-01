package datadog

import (
	"fmt"
	"regexp"
	"testing"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogIntegrationAwsTagFilter_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	uniqueID := uniqueAWSAccountID(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationAwsTagFilterDestroy(accProvider, "datadog_integration_aws_tag_filter.testing_aws_tag_filter"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAwsTagFilter_Basic(uniqueID),
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

func testAccCheckDatadogIntegrationAwsTagFilter_Basic(uniq string) string {
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

func testAccCheckDatadogIntegrationAwsTagFilterExists(accProvider *schema.Provider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceId := s.RootModule().Resources[resourceName].Primary.ID
		_, tfNamespace, err := accountAndNamespaceFromID(resourceId)
		namespace := datadogV1.AWSNamespace(tfNamespace)

		filters, err := listFiltersHelper(accProvider, resourceId)
		if err != nil {
			return err
		}

		for _, filter := range filters {
			if filter.GetNamespace() == namespace {
				if len(filter.GetTagFilterStr()) == 0 {
					return translateClientError(nil, fmt.Sprintf("tag_filter_str is empty for resource %s", namespace))
				}
				return nil
			}
		}

		return nil
	}
}

func testAccCheckDatadogIntegrationAwsTagFilterDestroy(accProvider *schema.Provider, resourceName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		resourceId := s.RootModule().Resources[resourceName].Primary.ID
		_, tfNamespace, err := accountAndNamespaceFromID(resourceId)
		namespace := datadogV1.AWSNamespace(tfNamespace)

		filters, err := listFiltersHelper(accProvider, resourceId)
		if err != nil {
			if matched, _ := regexp.MatchString("AWS account [0-9]+ does not exist in integration", err.Error()); matched {
				return nil
			}
			return err
		}

		for _, filter := range filters {
			if filter.GetNamespace() == namespace {
				if len(filter.GetTagFilterStr()) != 0 {
					return translateClientError(nil, fmt.Sprintf("tag_filter_str is not empty for namespace %s", namespace))
				}
				return nil
			}
		}

		return nil
	}
}

func listFiltersHelper(accProvider *schema.Provider, resourceId string) ([]datadogV1.AWSTagFilterListResponseFilters, error) {
	meta := accProvider.Meta()
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	filters := []datadogV1.AWSTagFilterListResponseFilters{}
	accountID, _, err := accountAndNamespaceFromID(resourceId)
	if err != nil {
		return nil, err
	}

	resp, _, err := datadogClient.AWSIntegrationApi.ListAWSTagFilters(auth).AccountId(accountID).Execute()
	if err != nil {
		return nil, err
	}
	filters = append(filters, resp.GetFilters()...)

	return filters, nil
}
