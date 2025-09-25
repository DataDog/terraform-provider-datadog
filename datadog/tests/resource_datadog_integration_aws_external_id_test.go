package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccIntegrationAwsExternalIDBasic(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAwsExternalID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("datadog_integration_aws_external_id.foo", "id"),
					testAccCheckDatadogIntegrationAwsExternalID_Create(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationAwsExternalID() string {
	return `resource "datadog_integration_aws_external_id" "foo" {}`
}

func testAccCheckDatadogIntegrationAwsExternalID_Create(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := IntegrationAwsExternalIDCreateHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationAwsExternalIDCreateHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_aws_external_id" {
				continue
			}

			_, httpResp, err := apiInstances.GetAWSIntegrationApiV2().CreateNewAWSExternalID(auth)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error creating external id")
			}
		}
		return nil
	})
	return err
}
