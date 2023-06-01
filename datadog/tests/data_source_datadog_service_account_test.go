package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogServiceAccountDatasourceExactMatch(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	serviceAccountName := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceAccountConfig(serviceAccountName),
				Check:  resource.TestCheckResourceAttr("data.datadog_service_account.test", "email", serviceAccountName),
			},
		},
	})
}

func TestAccDatadogServiceAccountDatasourceError(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	serviceAccountName := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceServiceAccountError(),
				ExpectError: regexp.MustCompile("didn't find any service account user matching"),
			},
			{
				Config:      testAccDatasourceServiceAccountHumanUser(serviceAccountName),
				ExpectError: regexp.MustCompile("your query returned a human user and not a service account user"),
			},
		},
	})
}

func testAccDatasourceServiceAccountConfig(uniq string) string {
	return fmt.Sprintf(`
	data "datadog_service_account" "test" {
	  filter = "%s"
	  depends_on = [
	    datadog_service_account.foo
	  ]
	}
	resource "datadog_service_account" "foo" {
    email = "%s"
    }`, uniq, uniq)
}

func testAccDatasourceServiceAccountHumanUser(uniq string) string {
	return fmt.Sprintf(`
	data "datadog_service_account" "test" {
	  filter = "%s"
	  depends_on = [
	    datadog_user.foo
	  ]
	}
	resource "datadog_user" "foo" {
    email = "%s"
    }`, uniq, uniq)
}

func testAccDatasourceServiceAccountError() string {
	return `	
	data "datadog_service_account" "test" {
		filter = "doesntexist01b0bb82-2000-4113-bb6b-e34b48ef37ff@example.com'"
	}
`
}
