package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogUserDatasourceExactMatch(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				// Step 1: Create user first to allow time for API indexing
				Config: testAccDatasourceUserConfigUserOnly(username),
				Check:  resource.TestCheckResourceAttr("datadog_user.foo", "email", username),
			},
			{
				// Step 2: Wait for API indexing, then add the data source lookup
				PreConfig: func() { time.Sleep(5 * time.Second) },
				Config:    testAccDatasourceUserConfig(username),
				Check:     resource.TestCheckResourceAttr("data.datadog_user.test", "email", username),
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				// Step 1: Create users first to allow time for API indexing
				Config: testAccDatasourceUserWithExactMatchConfigUsersOnly(email),
				Check:  resource.TestCheckResourceAttr("datadog_user.foo", "email", email),
			},
			{
				// Step 2: Wait for API indexing, then add the data source lookup
				PreConfig: func() { time.Sleep(5 * time.Second) },
				Config:    testAccDatasourceUserWithExactMatchConfig(email, "true"),
				Check:     resource.TestCheckResourceAttr("data.datadog_user.test", "email", email),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExactMatchError(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				// Step 1: Create users first to allow time for API indexing
				Config: testAccDatasourceUserWithExactMatchConfigUsersOnly(email),
				Check:  resource.TestCheckResourceAttr("datadog_user.foo", "email", email),
			},
			{
				// Step 2: Wait for API indexing, then add the data source lookup (expect error for multiple results)
				PreConfig:   func() { time.Sleep(5 * time.Second) },
				Config:      testAccDatasourceUserWithExactMatchConfig(email, "false"),
				ExpectError: regexp.MustCompile("your query returned more than one result for filter"),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExcludeServiceAccounts(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	serviceAccountName := ("tf-service-account-" + strings.ToLower(uniqueEntityName(ctx, t)))[:50]
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Step 1: Create user and service account first to allow time for API indexing
				Config: testAccDatasourceUserWithExcludeServiceAccountsConfigResourcesOnly(email, serviceAccountName),
				Check:  resource.TestCheckResourceAttr("datadog_user.foo", "email", email),
			},
			{
				// Step 2: Wait for API indexing, then add the data source lookup
				PreConfig: func() { time.Sleep(5 * time.Second) },
				Config:    testAccDatasourceUserWithExcludeServiceAccountsConfig(email, serviceAccountName, "true"),
				Check:     resource.TestCheckResourceAttr("data.datadog_user.test", "email", email),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExcludeServiceAccountsWithError(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	serviceAccountName := ("tf-service-account-" + strings.ToLower(uniqueEntityName(ctx, t)))[:50]
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Step 1: Create user and service account first to allow time for API indexing
				Config: testAccDatasourceUserWithExcludeServiceAccountsConfigResourcesOnly(email, serviceAccountName),
				Check:  resource.TestCheckResourceAttr("datadog_user.foo", "email", email),
			},
			{
				// Step 2: Wait for API indexing, then add the data source lookup (expect error for multiple results)
				PreConfig:   func() { time.Sleep(5 * time.Second) },
				Config:      testAccDatasourceUserWithExcludeServiceAccountsConfig(email, serviceAccountName, "false"),
				ExpectError: regexp.MustCompile("your query returned more than one result for filter"),
			},
		},
	})
}

func TestAccDatadogUserDatasourceWithExcludeServiceAccountsMultipleUsersWithError(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	serviceAccountName := ("tf-service-account-" + strings.ToLower(uniqueEntityName(ctx, t)))[:50]
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Step 1: Create users and service account first to allow time for API indexing
				Config: testAccDatasourceUserWithExcludeServiceAccountsMultipleUsersConfigResourcesOnly(email, serviceAccountName),
				Check:  resource.TestCheckResourceAttr("datadog_user.foo", "email", email),
			},
			{
				// Step 2: Wait for API indexing, then add the data source lookup (expect error for multiple users)
				PreConfig:   func() { time.Sleep(5 * time.Second) },
				Config:      testAccDatasourceUserWithExcludeServiceAccountsMultipleUsersConfig(email, serviceAccountName, "true"),
				ExpectError: regexp.MustCompile("after excluding service accounts, your query returned more than one result for filter"),
			},
		},
	})
}

func testAccDatasourceUserWithExactMatchConfigUsersOnly(uniq string) string {
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
	return fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%[1]s"
	exact_match = %[2]s
	depends_on = [ datadog_user.foo, datadog_user.bar ]
}

resource "datadog_user" "foo" {
	email = "%[1]s"
	send_user_invitation = false
}
resource "datadog_user" "bar" {
	email = "other%[1]s"
	send_user_invitation = false
}`, uniq, exactMatch)
}

func testAccDatasourceUserWithExcludeServiceAccountsConfigResourcesOnly(uniq, serviceAccountName string) string {
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
	return fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%[1]s"
	exclude_service_accounts = %[3]s
	depends_on = [ datadog_user.foo, datadog_service_account.bar ]
}

resource "datadog_user" "foo" {
	email = "%[1]s"
	send_user_invitation = false
}
resource "datadog_service_account" "bar" {
	email = "%[1]s"
	name = "%[2]s"
}`, uniq, serviceAccountName, excludeServiceAccounts)
}

func testAccDatasourceUserWithExcludeServiceAccountsMultipleUsersConfigResourcesOnly(uniq, serviceAccountName string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
	email = "%[1]s"
	send_user_invitation = false
}
resource "datadog_user" "otherfoo" {
	email = "other%[1]s"
	send_user_invitation = false
}
resource "datadog_service_account" "bar" {
	email = "%[1]s"
	name = "%[2]s"
}`, uniq, serviceAccountName)
}

func testAccDatasourceUserWithExcludeServiceAccountsMultipleUsersConfig(uniq, serviceAccountName, excludeServiceAccounts string) string {
	return fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%[1]s"
	exclude_service_accounts = %[3]s
	depends_on = [ datadog_user.foo, datadog_service_account.bar ]
}

resource "datadog_user" "foo" {
	email = "%[1]s"
	send_user_invitation = false
}
resource "datadog_user" "otherfoo" {
	email = "other%[1]s"
	send_user_invitation = false
}
resource "datadog_service_account" "bar" {
	email = "%[1]s"
	name = "%[2]s"
}`, uniq, serviceAccountName, excludeServiceAccounts)
}

func testAccDatasourceUserConfigUserOnly(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
    email = "%s"
    send_user_invitation = false
}`, uniq)
}

func testAccDatasourceUserConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_user" "test" {
	filter = "%s"
	depends_on = [
		datadog_user.foo
	]
}
resource "datadog_user" "foo" {
    email = "%s"
    send_user_invitation = false
}`, uniq, uniq)
}

func testAccDatasourceUserError() string {
	return `
	data "datadog_user" "test" {
		filter = "doesntexist01b0bb82-2000-4113-bb6b-e34b48ef37ff@example.com'"
	}
`
}
