package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

var (
	testAWSAccountID = "123456789012"
	testAWSRole      = "role"
)

func TestAccDatadogActionConnectionDatasource_AWS_AssumeRole(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, "datadog_action_connection.aws_assume_role_conn"),
		Steps: []resource.TestStep{
			{
				Config: testAWSAssumeRoleConnectionDataSourceConfig(connectionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_action_connection.aws_assume_role_conn", "name", connectionName),
					resource.TestCheckResourceAttr("data.datadog_action_connection.aws_assume_role_conn", "aws.assume_role.account_id", testAWSAccountID),
					resource.TestCheckResourceAttr("data.datadog_action_connection.aws_assume_role_conn", "aws.assume_role.role", testAWSRole),
					resource.TestCheckResourceAttrSet("data.datadog_action_connection.aws_assume_role_conn", "aws.assume_role.principal_id"),
					resource.TestCheckResourceAttrSet("data.datadog_action_connection.aws_assume_role_conn", "aws.assume_role.external_id"),
				),
			},
		},
	})
}

func testAWSAssumeRoleConnectionDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	%s
	data "datadog_action_connection" "aws_assume_role_conn" {
		id = datadog_action_connection.aws_assume_role_conn.id
		depends_on = [datadog_action_connection.aws_assume_role_conn]
	}`, testAWSAssumeRoleConnectionResourceConfig(name))
}

func testAWSAssumeRoleConnectionResourceConfig(name string) string {
	return fmt.Sprintf(`
	resource "datadog_action_connection" "aws_assume_role_conn" {
		name = "%s"

		aws {
			assume_role {
				account_id = "%s"
				role = "%s"
			}
		}
	}`, name, testAWSAccountID, testAWSRole)
}

func testAccCheckDatadogConnectionDestroy(accProvider *fwprovider.FrameworkProvider, resourceName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resource := s.RootModule().Resources[resourceName]
		_, httpRes, err := apiInstances.GetActionConnectionApiV2().GetActionConnection(auth, resource.Primary.ID)
		if err != nil {
			if httpRes.StatusCode == 404 {
				return nil
			}
			return err
		}

		return fmt.Errorf("connection destroy check failed")
	}
}
