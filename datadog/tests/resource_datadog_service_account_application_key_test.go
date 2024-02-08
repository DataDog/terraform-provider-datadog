package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccServiceAccountApplicationKeyBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	uniqUpdated := uniq + "updated"
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogServiceAccountApplicationKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: generateServiceAccountApplicationKeyConfig("some_linked_users_email@test.com", "Service account linked to some user", uniq, []string{"data.datadog_role.ro_role.id"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceAccountApplicationKeyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_service_account_application_key.foo", "name", uniq),
					resource.TestCheckResourceAttrSet(
						"datadog_service_account_application_key.foo", "key"),
					resource.TestCheckResourceAttrSet(
						"datadog_service_account_application_key.foo", "created_at"),
					resource.TestCheckResourceAttrSet(
						"datadog_service_account_application_key.foo", "last4"),
					resource.TestCheckResourceAttrPair(
						"datadog_service_account_application_key.foo", "service_account_id", "datadog_service_account.bar", "id"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.bar", "roles.#", "1"),
				),
			},
			{
				Config: generateServiceAccountApplicationKeyConfig("some_linked_users_email@test.com", "Service account linked to some user", uniqUpdated, []string{"data.datadog_role.ro_role.id"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceAccountApplicationKeyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_service_account_application_key.foo", "name", uniqUpdated),
					resource.TestCheckResourceAttrSet(
						"datadog_service_account_application_key.foo", "key"),
					resource.TestCheckResourceAttrSet(
						"datadog_service_account_application_key.foo", "created_at"),
					resource.TestCheckResourceAttrSet(
						"datadog_service_account_application_key.foo", "last4"),
					resource.TestCheckResourceAttrPair(
						"datadog_service_account_application_key.foo", "service_account_id", "datadog_service_account.bar", "id"),
					resource.TestCheckResourceAttr(
						"datadog_service_account.bar", "roles.#", "1"),
				),
			},
		},
	})
}

func TestAccServiceAccountApplicationKeyBasic_import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_service_account_application_key.foo"
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogServiceAccountApplicationKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: generateServiceAccountApplicationKeyConfig("some_linked_users_email@test.com", "Service account linked to some user", uniq, []string{"data.datadog_role.ro_role.id"}),
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

func generateServiceAccountApplicationKeyConfig(email, name, uniq string, roles []string) string {
	return fmt.Sprintf(`
        data "datadog_role" "ro_role" {
          filter = "Datadog Read Only Role"
        }

        resource "datadog_service_account" "bar" {
          email = "%v"
          name = "%v"
          roles = [%v]
        }

		resource "datadog_service_account_application_key" "foo" {
			service_account_id = datadog_service_account.bar.id
			name = "%s"
		}`, email, name, strings.Join(roles, ","), uniq)
}

func testAccCheckDatadogServiceAccountApplicationKeyDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := ServiceAccountApplicationKeyDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func ServiceAccountApplicationKeyDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_service_account_application_key" {
				continue
			}
			serviceAccountId := r.Primary.Attributes["service_account_id"]
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetServiceAccountsApiV2().GetServiceAccountApplicationKey(auth, serviceAccountId, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving ServiceAccountApplicationKey %s", err)}
			}
			return &utils.RetryableError{Prob: "ServiceAccountApplicationKey still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogServiceAccountApplicationKeyExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := serviceAccountApplicationKeyExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func serviceAccountApplicationKeyExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_service_account_application_key" {
			continue
		}
		serviceAccountId := r.Primary.Attributes["service_account_id"]
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetServiceAccountsApiV2().GetServiceAccountApplicationKey(auth, serviceAccountId, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving ServiceAccountApplicationKey")
		}
	}
	return nil
}
