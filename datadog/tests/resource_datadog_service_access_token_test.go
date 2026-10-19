package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccServiceAccessTokenBasic(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	uniqUpdated := uniq + "updated"
	scopes := []string{"dashboards_read", "dashboards_write"}
	updatedScopes := []string{"dashboards_read"}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogServiceAccessTokenDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceAccessTokenScoped(uniq, scopes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceAccessTokenExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_service_access_token.foo", "name", uniq),
					resource.TestCheckResourceAttrSet(
						"datadog_service_access_token.foo", "key"),
					resource.TestCheckResourceAttrSet(
						"datadog_service_access_token.foo", "created_at"),
					resource.TestCheckResourceAttrSet(
						"datadog_service_access_token.foo", "public_portion"),
					resource.TestCheckResourceAttrPair(
						"datadog_service_access_token.foo", "service_account_id", "datadog_service_account.bar", "id"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_access_token.foo", "scopes.*", scopes[0]),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_access_token.foo", "scopes.*", scopes[1]),
				),
			},
			{
				Config: testAccCheckDatadogServiceAccessTokenScoped(uniqUpdated, scopes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceAccessTokenExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_service_access_token.foo", "name", uniqUpdated),
					resource.TestCheckResourceAttrSet(
						"datadog_service_access_token.foo", "key"),
				),
			},
			{
				Config: testAccCheckDatadogServiceAccessTokenScoped(uniqUpdated, updatedScopes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceAccessTokenExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_service_access_token.foo", "name", uniqUpdated),
					resource.TestCheckTypeSetElemAttr(
						"datadog_service_access_token.foo", "scopes.*", updatedScopes[0]),
				),
			},
		},
	})
}

func TestAccServiceAccessToken_Error(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	tokenName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogServiceAccessTokenScoped(tokenName, []string{}),
				ExpectError: regexp.MustCompile(`Attribute scopes set must contain at least 1 elements`),
			},
			{
				Config:      testAccCheckDatadogServiceAccessTokenScoped(tokenName, []string{"invalid"}),
				ExpectError: regexp.MustCompile(`(?i)invalid scopes`),
			},
		},
	})
}

func TestAccServiceAccessToken_EmptyName(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogServiceAccessTokenScoped("", []string{"dashboards_read"}),
				ExpectError: regexp.MustCompile(`(?i)string length must be at least 1`),
			},
		},
	})
}

func TestAccServiceAccessToken_Import(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	resourceName := "datadog_service_access_token.foo"
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogServiceAccessTokenDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceAccessTokenScoped(uniq, []string{"dashboards_read"}),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					resources := state.RootModule().Resources
					resourceState := resources[resourceName]
					return resourceState.Primary.Attributes["service_account_id"] + ":" + resourceState.Primary.Attributes["id"], nil
				},
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"},
			},
		},
	})
}

func formatScopes(scopes []string) string {
	quoted := make([]string, len(scopes))
	for i, s := range scopes {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func testAccCheckDatadogServiceAccessTokenScoped(uniq string, scopes []string) string {
	return fmt.Sprintf(`
resource "datadog_service_account" "bar" {
	email = "new@example.com"
	name  = "testTerraformServiceAccessTokens"
}

resource "datadog_service_access_token" "foo" {
    service_account_id = datadog_service_account.bar.id
    name = %q
	scopes = %s
}`, uniq, formatScopes(scopes))
}

func testAccCheckDatadogServiceAccessTokenWithExpiry(uniq string, scopes []string, expiresAt string) string {
	return fmt.Sprintf(`
resource "datadog_service_account" "bar" {
	email = "new@example.com"
	name  = "testTerraformServiceAccessTokens"
}

resource "datadog_service_access_token" "foo" {
    service_account_id = datadog_service_account.bar.id
    name = %q
	scopes = %s
	expires_at = %q
}`, uniq, formatScopes(scopes), expiresAt)
}

func TestAccServiceAccessToken_AddExpiryFails(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	expires := time.Now().UTC().Add(30 * 24 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogServiceAccessTokenDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Step 1: create without expires_at.
			{
				Config: testAccCheckDatadogServiceAccessTokenScoped(uniq, []string{"dashboards_read"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceAccessTokenExists(providers.frameworkProvider),
				),
			},
			// Step 2: try to add expires_at — must fail (immutable).
			{
				Config:      testAccCheckDatadogServiceAccessTokenWithExpiry(uniq, []string{"dashboards_read"}, expires),
				ExpectError: regexp.MustCompile(`(?i)expires_at cannot be modified after creation`),
			},
		},
	})
}

func TestAccServiceAccessToken_WithExpiry(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	// API caps expirations at 365 days from now; use ~30 / ~60 days to stay well inside.
	expires := time.Now().UTC().Add(30 * 24 * time.Hour).Format(time.RFC3339)
	expiresChanged := time.Now().UTC().Add(60 * 24 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogServiceAccessTokenDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Step 1: create with expires_at — succeeds.
			{
				Config: testAccCheckDatadogServiceAccessTokenWithExpiry(uniq, []string{"dashboards_read"}, expires),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceAccessTokenExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_service_access_token.foo", "expires_at", expires),
				),
			},
			// Step 2: try to change expires_at — must fail at plan time (immutable).
			{
				Config:      testAccCheckDatadogServiceAccessTokenWithExpiry(uniq, []string{"dashboards_read"}, expiresChanged),
				ExpectError: regexp.MustCompile(`(?i)expires_at cannot be modified after creation`),
			},
			// Step 3: try to drop expires_at — also fails.
			{
				Config:      testAccCheckDatadogServiceAccessTokenScoped(uniq, []string{"dashboards_read"}),
				ExpectError: regexp.MustCompile(`(?i)expires_at cannot be modified after creation`),
			},
		},
	})
}

func testAccCheckDatadogServiceAccessTokenDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return ServiceAccessTokenDestroyHelper(auth, s, apiInstances)
	}
}

func ServiceAccessTokenDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	return utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_service_access_token" {
				continue
			}
			serviceAccountId := r.Primary.Attributes["service_account_id"]
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetServiceAccountsApiV2().GetServiceAccountAccessToken(auth, serviceAccountId, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving ServiceAccessToken %s", err)}
			}
			return &utils.RetryableError{Prob: "ServiceAccessToken still exists"}
		}
		return nil
	})
}

func testAccCheckDatadogServiceAccessTokenExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return serviceAccessTokenExistsHelper(auth, s, apiInstances)
	}
}

func serviceAccessTokenExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_service_access_token" {
			continue
		}
		serviceAccountId := r.Primary.Attributes["service_account_id"]
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetServiceAccountsApiV2().GetServiceAccountAccessToken(auth, serviceAccountId, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving ServiceAccessToken")
		}
	}
	return nil
}
