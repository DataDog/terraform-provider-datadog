package test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// githubOidcReplayAccountIdentifier is the Datadog identity the recorded mapping resolves to
// in the cassette. The live and record paths resolve the caller through
// datadog_current_user: the API requires the caller's permissions to be a superset of
// the mapped account's, which mapping the caller to itself guarantees. After
// recording, the cassette is scrubbed — the current_user interactions are removed
// and the recorded account identifiers are replaced with this synthetic handle — so
// the recording identity never lands in the repository. The literal cannot be used
// live; keep it in sync with the account_identifier recorded in the cassette.
const githubOidcReplayAccountIdentifier = "tf-github-oidc-replay-user@example.com"

const (
	// The sub owner must be a GitHub organization Datadog controls: a live
	// acceptance run creates a mapping that grants the mapped test identity to
	// any workload whose OIDC token matches, so a third-party owner would let
	// that owner authenticate as the mapped test user while the mapping exists.
	// The repository deliberately does not exist and must stay that way: the
	// API does not validate repository existence, and a nonexistent repository
	// means no real GitHub workload can ever mint a matching OIDC token.
	githubOidcTestSubClaim        = "repo:DataDog/terraform-provider-wif-acceptance-nonexistent:ref:refs/heads/main"
	githubOidcTestRepositoryClaim = "^DataDog/terraform-provider-wif-acceptance-nonexistent$"
)

func TestAccGithubOidcIdentityMapping(t *testing.T) {
	t.Parallel()
	skipIfNoCassette(t)
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_github_oidc_identity_mapping.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckGithubOidcIdentityMappingDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccGithubOidcIdentityMappingConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubOidcIdentityMappingExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "claim_matchers.sub", githubOidcTestSubClaim),
					resource.TestCheckResourceAttr(resourceName, "claim_matchers.repository", githubOidcTestRepositoryClaim),
					resource.TestCheckResourceAttrSet(resourceName, "account_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "account_uuid"),
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

func TestAccGithubOidcIdentityMappingServiceAccount(t *testing.T) {
	skipIfNoCassette(t)
	// Run before parallel tests: a mapping for the caller itself can prevent
	// the same caller from creating another mapping, even for a service account.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_github_oidc_identity_mapping.test"
	serviceAccountName := "datadog_service_account.test"
	serviceAccountConfig := fmt.Sprintf(`
data "datadog_role" "read_only" {
  filter = "Datadog Read Only Role"
}

resource "datadog_service_account" "test" {
  email = %q
  name  = "GitHub OIDC acceptance test"
  # The API requires a nonempty permission set contained in the caller's.
  roles = [data.datadog_role.read_only.id]
}
`, strings.ToLower(uniqueEntityName(ctx, t))+"@datadoghq.com")
	// The service account email domain must satisfy org login-domain allowlists
	// (staging orgs typically admit only @datadoghq.com) and must stay hardcoded:
	// the email is embedded in recorded request bodies, so a domain that differs
	// between record and replay would break cassette body matching. The email also
	// embeds t.Name() through uniqueEntityName: renaming this test changes the
	// email replay generates, so the recorded email in the cassette must be
	// updated to match (or the cassette re-recorded).

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckGithubOidcIdentityMappingDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccGithubOidcIdentityMappingConfigWithIdentity(serviceAccountName+".id", serviceAccountConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGithubOidcIdentityMappingExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "account_identifier", serviceAccountName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_uuid", serviceAccountName, "id"),
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

func testAccGithubOidcIdentityMappingConfig() string {
	currentUserConfig := `data "datadog_current_user" "test" {}`
	accountIdentifier := "data.datadog_current_user.test.handle"
	if isReplaying() {
		currentUserConfig = ""
		accountIdentifier = fmt.Sprintf("%q", githubOidcReplayAccountIdentifier)
	}

	return testAccGithubOidcIdentityMappingConfigWithIdentity(accountIdentifier, currentUserConfig)
}

func testAccGithubOidcIdentityMappingConfigWithIdentity(accountIdentifier, identityConfig string) string {
	return fmt.Sprintf(`
%[2]s

resource "datadog_github_oidc_identity_mapping" "test" {
  account_identifier = %[1]s
  claim_matchers = {
    sub        = %[3]q
    repository = %[4]q
  }
}
`, accountIdentifier, identityConfig, githubOidcTestSubClaim, githubOidcTestRepositoryClaim)
}

func testAccCheckGithubOidcIdentityMappingExists(provider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		stateResource, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		return retry.RetryContext(provider.Auth, 30*time.Second, func() *retry.RetryError {
			_, httpResponse, err := provider.DatadogApiInstances.GetCloudAuthenticationApiV2().GetGitHubCloudAuthPersonaMapping(provider.Auth, stateResource.Primary.ID)
			if err != nil {
				if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
					return retry.RetryableError(utils.TranslateClientError(err, httpResponse, "error checking GitHub OIDC identity mapping existence"))
				}
				return retry.NonRetryableError(utils.TranslateClientError(err, httpResponse, "error checking GitHub OIDC identity mapping existence"))
			}
			return nil
		})
	}
}

func testAccCheckGithubOidcIdentityMappingDestroy(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, stateResource := range state.RootModule().Resources {
			if stateResource.Type != "datadog_github_oidc_identity_mapping" {
				continue
			}

			err := retry.RetryContext(provider.Auth, 30*time.Second, func() *retry.RetryError {
				_, httpResponse, err := provider.DatadogApiInstances.GetCloudAuthenticationApiV2().GetGitHubCloudAuthPersonaMapping(provider.Auth, stateResource.Primary.ID)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
						return nil
					}
					return retry.NonRetryableError(utils.TranslateClientError(err, httpResponse, "error checking GitHub OIDC identity mapping destruction"))
				}
				return retry.RetryableError(fmt.Errorf("GitHub OIDC identity mapping %s still exists", stateResource.Primary.ID))
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
