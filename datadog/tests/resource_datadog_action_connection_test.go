package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	testConnectionBaseURL = "https://catfact.ninja"
)

func TestAccDatadogActionConnectionResource_AWS_AssumeRole(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_action_connection.aws_assume_role_conn"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAWSAssumeRoleConnectionResourceConfig(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogConnectionExists(providers.frameworkProvider, resourceName),
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

func TestAccDatadogActionConnectionResource_HTTP_TokenAuth(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying - requires actions API access permission on API key")
	}
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_action_connection.http_token_auth_conn"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testHTTPTokenAuthConnectionResourceConfig(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogConnectionExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "http.base_url", testConnectionBaseURL),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.token.0.type", "SECRET"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.token.0.name", "token1"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.token.0.value", "secret value 1"), // retrieved from state
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.header.0.name", "header-name"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.header.0.value", "header-value"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.url_parameter.0.name", "urlParamName"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.url_parameter.0.value", "urlParamValue"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.body.content_type", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.body.content", "{\"key\":\"{{ token1 }}\"}"),
				),
			},
		},
	})
}

func testHTTPTokenAuthConnectionResourceConfig(name string) string {
	return fmt.Sprintf(`
	resource "datadog_action_connection" "http_token_auth_conn" {
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

func testAccCheckDatadogConnectionExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogConnectionExistsHelper(auth, s, apiInstances, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogConnectionExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	id := s.RootModule().Resources[name].Primary.ID
	if _, _, err := apiInstances.GetActionConnectionApiV2().GetActionConnection(ctx, id); err != nil {
		return fmt.Errorf("received an error retrieving connection: %s", err)
	}
	return nil
}
