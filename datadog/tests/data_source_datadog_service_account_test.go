package test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogServiceAccountDatasourceMatchFilter(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying - eventual consistency with service account creation")
	}
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	unique_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniqueEntityName(ctx, t))))
	name := fmt.Sprintf("tf-TestAccServiceAccountDatasource-%s", unique_hash[:16])
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
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

func testAccDatasourceServiceAccountConfig(email string, name string) string {
	return fmt.Sprintf(`
	data "datadog_service_account" "test" {
		filter = "%[2]s"
		depends_on = [
			datadog_service_account.foo
		]
	}
	resource "datadog_service_account" "foo" {
		email = "%[1]s"
		name = "%[2]s"
	}`, email, name)
}

func TestAccDatadogServiceAccountDatasourceMatchID(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	unique_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniqueEntityName(ctx, t))))
	name := fmt.Sprintf("tf-TestAccServiceAccountDatasource-%s", unique_hash[:16])
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
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
		ProtoV5ProviderFactories: accProviders,
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
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqueHash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniqueEntityName(ctx, t))))
	namePrefix := fmt.Sprintf("tf-TestAccServiceAccountPagination-%s", uniqueHash[:16])
	emailSuffix := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
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

func testAccDatasourceServiceAccountPaginationConfig(emailSuffix, namePrefix string) string {
	return fmt.Sprintf(`
	resource "datadog_service_account" "test" {
		count = 20
		email = "${count.index + 1}-%[1]s"
		name  = "%[2]s-${count.index + 1}"
	}

	data "datadog_service_account" "foo" {
		filter = "%[2]s-20"
		depends_on = [datadog_service_account.test]
	}`, emailSuffix, namePrefix)
}
