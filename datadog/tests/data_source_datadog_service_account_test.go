package test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// testAccCheckServiceAccountSearchable polls the users list API until the
// filter returns a result. The users search index is eventually consistent,
// so a data source read in the same apply as the create can return no results.
func testAccCheckServiceAccountSearchable(provider *fwprovider.FrameworkProvider, filter string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		return retry.RetryContext(provider.Auth, 2*time.Minute, func() *retry.RetryError {
			resp, _, err := provider.DatadogApiInstances.GetUsersApiV2().ListUsers(provider.Auth, datadogV2.ListUsersOptionalParameters{Filter: &filter})
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("error listing users: %w", err))
			}
			if len(resp.GetData()) == 0 {
				return retry.RetryableError(fmt.Errorf("filter %q returned no results", filter))
			}
			return nil
		})
	}
}

func TestAccDatadogServiceAccountDatasourceMatchFilter(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	unique_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniqueEntityName(ctx, t))))
	name := fmt.Sprintf("tf-TestAccServiceAccountDatasource-%s", unique_hash[:16])
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceAccountResourceConfig(email, name),
				Check:  testAccCheckServiceAccountSearchable(providers.frameworkProvider, name),
			},
			{
				Config: testAccDatasourceServiceAccountConfig(email, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.datadog_service_account.test", "email", "datadog_service_account.foo", "email"),
					resource.TestCheckResourceAttrPair("data.datadog_service_account.test", "name", "datadog_service_account.foo", "name"),
				),
			},
		},
	})
}

func testAccDatasourceServiceAccountResourceConfig(email string, name string) string {
	return fmt.Sprintf(`
	resource "datadog_service_account" "foo" {
		email = "%[1]s"
		name = "%[2]s"
	}`, email, name)
}

func testAccDatasourceServiceAccountConfig(email string, name string) string {
	return testAccDatasourceServiceAccountResourceConfig(email, name) + fmt.Sprintf(`
	data "datadog_service_account" "test" {
		filter = "%[1]s"
		depends_on = [
			datadog_service_account.foo
		]
	}`, name)
}

func TestAccDatadogServiceAccountDatasourceMatchID(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	unique_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniqueEntityName(ctx, t))))
	name := fmt.Sprintf("tf-TestAccServiceAccountDatasource-%s", unique_hash[:16])
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceAccountConfigID(email, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.datadog_service_account.test", "email", "datadog_service_account.foo", "email"),
					resource.TestCheckResourceAttrPair("data.datadog_service_account.test", "name", "datadog_service_account.foo", "name"),
				),
			},
		},
	})
}

func testAccDatasourceServiceAccountConfigID(email string, name string) string {
	return fmt.Sprintf(`
	data "datadog_service_account" "test" {
		id = datadog_service_account.foo.id
	}
	resource "datadog_service_account" "foo" {
		email = "%s"
		name = "%s"
	}`, email, name)
}

func TestAccDatadogServiceAccountDatasourceError(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceServiceAccountNotFoundError(email),
				ExpectError: regexp.MustCompile("filter keyword returned no results"),
			},
			{
				Config:      testAccDatasourceServiceAccountUserError(email),
				ExpectError: regexp.MustCompile("Obtained entity was not a service account"),
			},
		},
	})
}

func testAccDatasourceServiceAccountUserError(email string) string {
	return fmt.Sprintf(`
	data "datadog_service_account" "test" {
		id = datadog_user.bar.id
	}
	resource "datadog_user" "bar" {
		email = "%s"
	}`, email)
}

func testAccDatasourceServiceAccountNotFoundError(uniq string) string {
	return fmt.Sprintf(`	
	data "datadog_service_account" "test" {
		filter = "does_not_exist_%s"
	}`, uniq)
}

func TestAccDatadogServiceAccountDatasourcePagination(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqueHash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniqueEntityName(ctx, t))))
	namePrefix := fmt.Sprintf("tf-TestAccServiceAccountPagination-%s", uniqueHash[:16])
	emailSuffix := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceAccountPaginationResourcesConfig(emailSuffix, namePrefix),
				Check:  testAccCheckServiceAccountSearchable(providers.frameworkProvider, fmt.Sprintf("%s-20", namePrefix)),
			},
			{
				Config: testAccDatasourceServiceAccountPaginationConfig(emailSuffix, namePrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_service_account.foo", "email", fmt.Sprintf("20-%s", emailSuffix)),
					resource.TestCheckResourceAttr("data.datadog_service_account.foo", "name", fmt.Sprintf("%s-20", namePrefix)),
				),
			},
		},
	})
}

func testAccDatasourceServiceAccountPaginationResourcesConfig(emailSuffix, namePrefix string) string {
	return fmt.Sprintf(`
	resource "datadog_service_account" "test" {
		count = 20
		email = "${count.index + 1}-%[1]s"
		name  = "%[2]s-${count.index + 1}"
	}`, emailSuffix, namePrefix)
}

func testAccDatasourceServiceAccountPaginationConfig(emailSuffix, namePrefix string) string {
	return testAccDatasourceServiceAccountPaginationResourcesConfig(emailSuffix, namePrefix) + fmt.Sprintf(`

	data "datadog_service_account" "foo" {
		filter = "%s-20"
		depends_on = [datadog_service_account.test]
	}`, namePrefix)
}
