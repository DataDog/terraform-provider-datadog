package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccServiceAccountApplicationKeyBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogServiceAccountApplicationKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceAccountApplicationKey(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogServiceAccountApplicationKeyExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(
						"datadog_service_account_application_key.foo", "name", "Application Key for managing dashboards"),
				),
			},
		},
	})
}

func testAccCheckDatadogServiceAccountApplicationKey(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_service_account_application_key" "foo" {
    service_account_id = "00000000-0000-1234-0000-000000000000"
    name = "Application Key for managing dashboards"
    scopes = ["dashboards_read", "dashboards_write", "dashboards_public_share"]
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
