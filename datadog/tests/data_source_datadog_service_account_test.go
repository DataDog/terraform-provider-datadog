package test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogServiceAccountDatasourceMatchFilter(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	unique_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniqueEntityName(ctx, t))))
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	name := fmt.Sprintf("tf-TestAccServiceAccountDatasource-%s", unique_hash[:16])
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogServiceAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceAccountConfig(email, name),
				Check:  resource.TestCheckResourceAttr("data.datadog_service_account.test", "email", email),
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
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	unique_hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniqueEntityName(ctx, t))))
	name := fmt.Sprintf("tf-TestAccServiceAccountDatasource-%s", unique_hash[:16])
	email := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogServiceAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceServiceAccountConfigID(email, name),
				Check:  resource.TestCheckResourceAttr("data.datadog_service_account.test", "email", email),
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

func testAccCheckDatadogServiceAccountDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogServiceAccountDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datadogServiceAccountDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_service_account" {
			continue
		}

		id := r.Primary.ID
		userResponse, httpResponse, err := apiInstances.GetUsersApiV2().GetUser(ctx, id)

		if err != nil {
			if httpResponse.StatusCode == 404 {
				continue
			}
			return fmt.Errorf("received an error retrieving service account %s", err)
		}
		userData := userResponse.GetData()
		userAttributes := userData.GetAttributes()
		// Datadog only disables user/service account on DELETE
		if userAttributes.GetDisabled() {
			continue
		}
		return fmt.Errorf("service account still exists")
	}

	return nil
}
