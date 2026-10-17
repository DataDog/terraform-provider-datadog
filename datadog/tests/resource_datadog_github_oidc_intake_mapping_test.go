package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccGithubOidcIntakeMapping(t *testing.T) {
	t.Parallel()
	skipIfNoCassette(t)
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_github_oidc_intake_mapping.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckGithubOidcIntakeMappingDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccGithubOidcIntakeMappingConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubOidcIntakeMappingExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "claim_matchers.sub", githubOidcTestSubClaim),
					resource.TestCheckResourceAttr(resourceName, "claim_matchers.repository", githubOidcTestRepositoryClaim),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGithubOidcIntakeMappingConfig() string {
	return fmt.Sprintf(`
resource "datadog_github_oidc_intake_mapping" "test" {
  claim_matchers = {
    sub        = %[1]q
    repository = %[2]q
  }
}
`, githubOidcTestSubClaim, githubOidcTestRepositoryClaim)
}

func testAccCheckGithubOidcIntakeMappingExists(provider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		stateResource, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		return retry.RetryContext(provider.Auth, 30*time.Second, func() *retry.RetryError {
			_, httpResponse, err := provider.DatadogApiInstances.GetCloudAuthenticationApiV2().GetGitHubCloudAuthIntakeMapping(provider.Auth, stateResource.Primary.ID)
			if err != nil {
				if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
					return retry.RetryableError(utils.TranslateClientError(err, httpResponse, "error checking GitHub OIDC intake mapping existence"))
				}
				return retry.NonRetryableError(utils.TranslateClientError(err, httpResponse, "error checking GitHub OIDC intake mapping existence"))
			}
			return nil
		})
	}
}

func testAccCheckGithubOidcIntakeMappingDestroy(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, stateResource := range state.RootModule().Resources {
			if stateResource.Type != "datadog_github_oidc_intake_mapping" {
				continue
			}

			err := retry.RetryContext(provider.Auth, 30*time.Second, func() *retry.RetryError {
				_, httpResponse, err := provider.DatadogApiInstances.GetCloudAuthenticationApiV2().GetGitHubCloudAuthIntakeMapping(provider.Auth, stateResource.Primary.ID)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
						return nil
					}
					return retry.NonRetryableError(utils.TranslateClientError(err, httpResponse, "error checking GitHub OIDC intake mapping destruction"))
				}
				return retry.RetryableError(fmt.Errorf("GitHub OIDC intake mapping %s still exists", stateResource.Primary.ID))
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
