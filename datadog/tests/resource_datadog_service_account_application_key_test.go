package test

import (
	"context"
	"fmt"
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
				Config: testAccCheckDatadogServiceAccountApplicationKey(uniq),
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
				),
			},
			{
				Config: testAccCheckDatadogServiceAccountApplicationKey(uniqUpdated),
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
				Config: testAccCheckDatadogServiceAccountApplicationKey(uniq),
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

func testAccCheckDatadogServiceAccountApplicationKey(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_account" "bar" {
	email = "new@example.com"
	name  = "testTerraformServiceAccountApplicationKeys"
}

resource "datadog_service_account_application_key" "foo" {
    service_account_id = datadog_service_account.bar.id
    name = "%s"
}`, uniq)
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
