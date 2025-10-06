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

func TestAccDatadogIntegrationAWSIAMPermissionsDatasource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAWSIAMPermissionsConfig(),
				Check: resource.ComposeTestCheckFunc(
					checkDatadogIntegrationAWSIAMPermissionsCount(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAWSIAMPermissionsConfig() string {
	return `data "datadog_integration_aws_iam_permissions" "foo" {}`
}

func checkDatadogIntegrationAWSIAMPermissionsCount(accProvider *fwprovider.FrameworkProvider) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		iamPermissions, _, err := apiInstances.GetAWSIntegrationApiV2().GetAWSIntegrationIAMPermissions(auth)
		if err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_integration_aws_iam_permissions.foo"].Primary.Attributes
		iamPermissionsCount, _ := strconv.Atoi(resourceAttributes["iam_permissions.#"])
		permissionsDd := iamPermissions.Data.Attributes.Permissions

		if iamPermissionsCount != len(permissionsDd) {
			return fmt.Errorf("expected %d iam permissions, got %d iam permissions",
				iamPermissionsCount, len(permissionsDd))
		}

		return nil
	}
}
