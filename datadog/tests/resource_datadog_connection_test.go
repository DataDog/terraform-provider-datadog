package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testConnectionBaseURL = "https://catfact.ninja"
)

func TestAccDatadogConnectionResource_AWS_AssumeRole(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_connection.aws_assume_role_conn"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAWSAssumeRoleConnectionResourceConfig(connectionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "aws.assume_role.account_id", testAWSAccountID),
					resource.TestCheckResourceAttr(resourceName, "aws.assume_role.role", testAWSRole),
					resource.TestCheckResourceAttrSet(resourceName, "aws.assume_role.principal_id"),
					resource.TestCheckResourceAttrSet(resourceName, "aws.assume_role.external_id"),
				),
			},
		},
	})
}

func TestAccDatadogConnectionResource_HTTP_TokenAuth(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_connection.http_token_auth_conn"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testHTTPTokenAuthConnectionResourceConfig(connectionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "http.base_url", testConnectionBaseURL),
				),
			},
		},
	})
}

func testHTTPTokenAuthConnectionResourceConfig(name string) string {
	return fmt.Sprintf(`
	resource "datadog_connection" "http_token_auth_conn" {
		name = "%s"

		http {
			base_url = "%s"

			token_auth {
				token {
					type = "SECRET"
					name = "token1"
					value = "secret value 1"
				}
				
				header {
					name = "header-name"
					value = "header-value"
				}

				url_parameter {
					name = "urlParamName"
					value = "urlParamValue"
				}

				body {
					content_type = "application/json"
					content = jsonencode({
						key = "{{ token1 }}"
					})
				}
			}
		}
	}`, name, testConnectionBaseURL)
}
