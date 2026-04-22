package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIntegrationAWSIAMPermissionsStandardDatasource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAWSIAMPermissionsStandardConfig(),
				Check: resource.ComposeTestCheckFunc(
					checkDatadogIntegrationAWSIAMPermissionsStandardCount(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAWSIAMPermissionsStandardConfig() string {
	return `data "datadog_integration_aws_iam_permissions_standard" "foo" {}`
}

func checkDatadogIntegrationAWSIAMPermissionsStandardCount(accProvider *fwprovider.FrameworkProvider) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		iamPermissions, _, err := apiInstances.GetAWSIntegrationApiV2().GetAWSIntegrationIAMPermissionsStandard(auth)
		if err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_integration_aws_iam_permissions_standard.foo"].Primary.Attributes
		iamPermissionsCount, _ := strconv.Atoi(resourceAttributes["iam_permissions.#"])
		permissionsDd := iamPermissions.Data.Attributes.Permissions

		if iamPermissionsCount != len(permissionsDd) {
			return fmt.Errorf("expected %d iam permissions, got %d iam permissions",
				iamPermissionsCount, len(permissionsDd))
		}

		return nil
	}
}
