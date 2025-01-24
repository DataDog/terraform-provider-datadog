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

func TestAccDatadogConnectionDatasource_AWS_AssumeRole(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testConnectionDataSourceConfig(connectionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_connection.conn", "name", connectionName),
					resource.TestCheckResourceAttr("datadog_connection.conn", "aws.assume_role.account_id", testAWSAccountID),
					resource.TestCheckResourceAttr("datadog_connection.conn", "aws.assume_role.role", testAWSRole),
					resource.TestCheckResourceAttrSet("datadog_connection.conn", "aws.assume_role.principal_id"),
					resource.TestCheckResourceAttrSet("datadog_connection.conn", "aws.assume_role.external_id"),
				),
			},
		},
	})
}

func testConnectionDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	%s
	data "datadog_connection" "conn" {
		id = datadog_connection.conn.id
		depends_on = [datadog_connection.conn]
	}`, testConnectionResourceConfig(name))
}

func testConnectionResourceConfig(name string) string {
	return fmt.Sprintf(`
	resource "datadog_connection" "conn" {
		name = "%s"

		aws {
			assume_role {
				account_id = "%s"
				role = "%s"
			}
		}
	}`, name, testAWSAccountID, testAWSRole)
}

func testAccCheckDatadogConnectionDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resource := s.RootModule().Resources["datadog_connection.conn"]
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
