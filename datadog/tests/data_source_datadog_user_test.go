package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogUserDatasourceExactMatch(t *testing.T) {
	parallelIfLive(t)
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserResourceConfig(username),
				Check:  testAccCheckUsersListed(sdkAPIClient(accProvider), username, "", 1),
			},
			{
				Config: testAccDatasourceUserConfig(username),
				Check:  resource.TestCheckResourceAttr("data.datadog_user.test", "email", username),
			},
		},
	})
}

func TestAccDatadogUserDatasourceError(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceUserError(),
				ExpectError: regexp.MustCompile("didn't find any user matching"),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExactMatch(t *testing.T) {
	parallelIfLive(t)
	ctx, accProviders := testAccProviders(context.Background(), t)
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserWithExactMatchResourcesConfig(email),
				Check:  testAccCheckUsersListed(sdkAPIClient(accProvider), email, "", 2),
			},
			{
				Config: testAccDatasourceUserWithExactMatchConfig(email, "true"),
				Check:  resource.TestCheckResourceAttr("data.datadog_user.test", "email", email),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExactMatchError(t *testing.T) {
	parallelIfLive(t)
	ctx, accProviders := testAccProviders(context.Background(), t)
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserWithExactMatchResourcesConfig(email),
				Check:  testAccCheckUsersListed(sdkAPIClient(accProvider), email, "", 2),
			},
			{
				Config:      testAccDatasourceUserWithExactMatchConfig(email, "false"),
				ExpectError: regexp.MustCompile("your query returned more than one result for filter"),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExcludeServiceAccounts(t *testing.T) {
	parallelIfLive(t)
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	serviceAccountName := ("tf-service-account-" + strings.ToLower(uniqueEntityName(ctx, t)))[:50]
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserWithExcludeServiceAccountsResourcesConfig(email, serviceAccountName),
				Check:  testAccCheckUsersListed(frameworkAPIClient(providers.frameworkProvider), email, "", 2),
			},
			{
				Config: testAccDatasourceUserWithExcludeServiceAccountsConfig(email, serviceAccountName, "true"),
				Check:  resource.TestCheckResourceAttr("data.datadog_user.test", "email", email),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExcludeServiceAccountsWithError(t *testing.T) {
	parallelIfLive(t)
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	serviceAccountName := ("tf-service-account-" + strings.ToLower(uniqueEntityName(ctx, t)))[:50]
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserWithExcludeServiceAccountsResourcesConfig(email, serviceAccountName),
				Check:  testAccCheckUsersListed(frameworkAPIClient(providers.frameworkProvider), email, "", 2),
			},
			{
				Config:      testAccDatasourceUserWithExcludeServiceAccountsConfig(email, serviceAccountName, "false"),
				ExpectError: regexp.MustCompile("your query returned more than one result for filter"),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExcludeServiceAccountsMultipleUsersWithError(t *testing.T) {
	parallelIfLive(t)
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	serviceAccountName := ("tf-service-account-" + strings.ToLower(uniqueEntityName(ctx, t)))[:50]
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceUserWithExcludeServiceAccountsMultipleUsersResourcesConfig(email, serviceAccountName),
				Check:  testAccCheckUsersListed(frameworkAPIClient(providers.frameworkProvider), email, "", 3),
			},
			{
				Config:      testAccDatasourceUserWithExcludeServiceAccountsMultipleUsersConfig(email, serviceAccountName, "true"),
				ExpectError: regexp.MustCompile("after excluding service accounts, your query returned more than one result for filter"),
			},
		},
	})
}

func testAccDatasourceUserWithExactMatchResourcesConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
	email = "%[1]s"
	send_user_invitation = false
}
resource "datadog_user" "bar" {
	email = "other%[1]s"
	send_user_invitation = false
}`, uniq)
}

func testAccDatasourceUserWithExactMatchConfig(uniq, exactMatch string) string {
	return testAccDatasourceUserWithExactMatchResourcesConfig(uniq) + fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%[1]s"
	exact_match = %[2]s
	depends_on = [ datadog_user.foo, datadog_user.bar ]
}`, uniq, exactMatch)
}

func testAccDatasourceUserWithExcludeServiceAccountsResourcesConfig(uniq, serviceAccountName string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
	email = "%[1]s"
	send_user_invitation = false
}
resource "datadog_service_account" "bar" {
	email = "%[1]s"
	name = "%[2]s"
}`, uniq, serviceAccountName)
}

func testAccDatasourceUserWithExcludeServiceAccountsConfig(uniq, serviceAccountName, excludeServiceAccounts string) string {
	return testAccDatasourceUserWithExcludeServiceAccountsResourcesConfig(uniq, serviceAccountName) + fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%[1]s"
	exclude_service_accounts = %[2]s
	depends_on = [ datadog_user.foo, datadog_service_account.bar ]
}`, uniq, excludeServiceAccounts)
}

func testAccDatasourceUserWithExcludeServiceAccountsMultipleUsersResourcesConfig(uniq, serviceAccountName string) string {
	return testAccDatasourceUserWithExcludeServiceAccountsResourcesConfig(uniq, serviceAccountName) + fmt.Sprintf(`
resource "datadog_user" "otherfoo" {
	email = "other%[1]s"
	send_user_invitation = false
}`, uniq)
}

func testAccDatasourceUserWithExcludeServiceAccountsMultipleUsersConfig(uniq, serviceAccountName, excludeServiceAccounts string) string {
	return testAccDatasourceUserWithExcludeServiceAccountsMultipleUsersResourcesConfig(uniq, serviceAccountName) + fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%[1]s"
	exclude_service_accounts = %[2]s
	depends_on = [ datadog_user.foo, datadog_user.otherfoo, datadog_service_account.bar ]
}`, uniq, excludeServiceAccounts)
}

func testAccDatasourceUserResourceConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
	email = "%s"
	send_user_invitation = false
}`, uniq)
}

func testAccDatasourceUserConfig(uniq string) string {
	return testAccDatasourceUserResourceConfig(uniq) + fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%s"
	depends_on = [
		datadog_user.foo
	]
}`, uniq)
}

func testAccDatasourceUserError() string {
	return `
	data "datadog_user" "test" {
		filter = "doesntexist01b0bb82-2000-4113-bb6b-e34b48ef37ff@example.com'"
	}
`
}
